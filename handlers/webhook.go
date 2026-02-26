package handlers

import (
	"alert-webhook/database"
	"alert-webhook/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

// HandleWebhook receives alerts from AlertManager and stores them in MySQL
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	var payload models.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("[错误] 解析Webhook请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Failed to parse webhook payload",
			"detail": err.Error(),
		})
		return
	}

	log.Printf("[收到] 接收到的告警数量: %d", len(payload.Alerts))
	log.Printf("[信息] Receiver: %s, GroupKey: %s", payload.Receiver, payload.GroupKey)

	db := database.GetDB()
	now := time.Now()
	alertsStored := 0

	for i, alertPayload := range payload.Alerts {
		// Extract severity from labels
		severity := alertPayload.Labels["severity"]

		var labels models.JSONMap = make(models.JSONMap)
		var annotations models.JSONMap = make(models.JSONMap)

		for k, v := range alertPayload.Labels {
			labels[k] = v
		}
		for k, v := range alertPayload.Annotations {
			annotations[k] = v
		}

		alert := models.Alert{
			Fingerprint:  alertPayload.Fingerprint,
			Alertname:    alertPayload.Labels["alertname"],
			Status:       models.AlertStatus(alertPayload.Status),
			Severity:     severity,
			StartsAt:     alertPayload.StartsAt,
			EndsAt:       alertPayload.EndsAt,
			CreatedAt:    now,
			Labels:       labels,
			Annotations:  annotations,
			GeneratorURL: alertPayload.GeneratorURL,
			Receiver:     payload.Receiver,
			GroupKey:     payload.GroupKey,
		}

		// Upsert: update if exists, insert if not
		result := db.Where("fingerprint = ?", alert.Fingerprint).First(&models.Alert{})

		if result.Error == gorm.ErrRecordNotFound {
			// Insert new alert
			if err := db.Create(&alert).Error; err != nil {
				log.Printf("[错误] 存储告警失败 [%d/%d]: %s - %v", i+1, len(payload.Alerts), alert.Alertname, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  "Failed to store alert",
					"detail": err.Error(),
				})
				return
			}
			log.Printf("[新增] 告警 [%d/%d]: %s | 状态: %s | 级别: %s", i+1, len(payload.Alerts), alert.Alertname, alert.Status, severity)
		} else if result.Error == nil {
			// Update existing alert
			updates := map[string]interface{}{
				"status":     alert.Status,
				"ends_at":    alert.EndsAt,
				"updated_at": now,
			}
			if err := db.Model(&alert).Where("fingerprint = ?", alert.Fingerprint).Updates(updates).Error; err != nil {
				log.Printf("[错误] 更新告警失败 [%d/%d]: %s - %v", i+1, len(payload.Alerts), alert.Alertname, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  "Failed to update alert",
					"detail": err.Error(),
				})
				return
			}
			log.Printf("[更新] 告警 [%d/%d]: %s | 状态: %s", i+1, len(payload.Alerts), alert.Alertname, alert.Status)
		} else {
			log.Printf("[错误] 数据库查询失败: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Database query failed",
				"detail": result.Error.Error(),
			})
			return
		}

		alertsStored++
	}

	log.Printf("[完成] 共处理 %d 条告警", alertsStored)

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"alerts_received": len(payload.Alerts),
		"alerts_stored":   alertsStored,
	})
}

// HealthCheck returns the health status of the service
func (h *WebhookHandler) HealthCheck(c *gin.Context) {
	db := database.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("[健康检查] 失败: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		log.Printf("[健康检查] 数据库连接失败: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
