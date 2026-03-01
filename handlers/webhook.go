package handlers

import (
	"alert-webhook/database"
	"alert-webhook/models"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

// HandleWebhook 接收来自 AlertManager 的告警，存入 MySQL
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	var payload models.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("[Webhook] 解析请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid payload",
			"detail": err.Error(),
		})
		return
	}

	log.Printf("[Webhook] 收到告警 %d 条, receiver=%s", len(payload.Alerts), payload.Receiver)

	db := database.GetDB()
	stored := 0

	for i, ap := range payload.Alerts {
		labels := models.JSONMap{}
		for k, v := range ap.Labels {
			labels[k] = v
		}
		annotations := models.JSONMap{}
		for k, v := range ap.Annotations {
			annotations[k] = v
		}

		alert := models.Alert{
			Fingerprint: ap.Fingerprint,
			Status:      models.AlertStatus(ap.Status),
			AlertName:   ap.Labels["alertname"],
			Instance:    ap.Labels["instance"],
			StartsAt:    ap.StartsAt,
			EndsAt:      ap.EndsAt,
			Labels:      labels,
			Annotations: annotations,
		}

		// 存储原始信息
		rawMap := make(map[string]interface{})
		rawBytes, _ := json.Marshal(ap)
		json.Unmarshal(rawBytes, &rawMap)
		alert.RawContent = rawMap

		// 根据状态执行不同逻辑
		if alert.Status == models.StatusFiring {
			// 避免直接使用 OnConflict (UPSERT) 导致 MySQL 自增主键产生空洞
			// 分为查询 -> 更新或插入 两步
			var existing models.Alert
			queryResult := db.Where("fingerprint = ? AND starts_at = ?", alert.Fingerprint, alert.StartsAt).First(&existing)
			
			if queryResult.Error == gorm.ErrRecordNotFound {
				// 不存在，创建新记录
				err := db.Create(&alert).Error
				if err != nil {
					log.Printf("[Webhook] Firing 存储失败 [%d/%d] %s: %v", i+1, len(payload.Alerts), alert.AlertName, err)
					continue
				}
			} else if queryResult.Error == nil {
				// 存在，更新记录
				err := db.Model(&existing).Updates(map[string]interface{}{
					"status":      alert.Status,
					"ends_at":     alert.EndsAt,
					"labels":      alert.Labels,
					"annotations": alert.Annotations,
					"raw_content": alert.RawContent,
				}).Error
				if err != nil {
					log.Printf("[Webhook] Firing 更新失败 [%d/%d] %s: %v", i+1, len(payload.Alerts), alert.AlertName, err)
					continue
				}
			} else {
				log.Printf("[Webhook] 查询记录失败 [%d/%d] %s: %v", i+1, len(payload.Alerts), alert.AlertName, queryResult.Error)
				continue
			}
		} else {
			// Resolved 状态：更新匹配的 firing 记录
			result := db.Model(&models.Alert{}).
				Where("fingerprint = ? AND starts_at = ?", alert.Fingerprint, alert.StartsAt).
				Updates(map[string]interface{}{
					"status":      models.StatusResolved,
					"ends_at":     alert.EndsAt,
					"raw_content": alert.RawContent,
				})

			if result.Error != nil {
				log.Printf("[Webhook] Resolved 更新失败 [%d/%d] %s: %v", i+1, len(payload.Alerts), alert.AlertName, result.Error)
				continue
			}

			// 如果没找到对应记录，则创建一条
			if result.RowsAffected == 0 {
				db.Create(&alert)
			}
		}

		log.Printf("[Webhook] 处理完成 [%d/%d] %s | %s", i+1, len(payload.Alerts), alert.AlertName, ap.Status)
		stored++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "ok",
		"alerts_received": len(payload.Alerts),
		"alerts_stored":   stored,
	})
}

// HealthCheck 服务健康检查
func (h *WebhookHandler) HealthCheck(c *gin.Context) {
	db := database.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
