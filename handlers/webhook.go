package handlers

import (
	"alert-webhook/database"
	"alert-webhook/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
	now := time.Now()
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

		alertname := ap.Labels["alertname"]
		severity := ap.Labels["severity"]

		alert := models.Alert{
			Fingerprint:  ap.Fingerprint,
			Alertname:    alertname,
			Status:       models.AlertStatus(ap.Status),
			Severity:     severity,
			Instance:     ap.Labels["instance"],
			Job:          ap.Labels["job"],
			Cluster:      ap.Labels["cluster"],
			Env:          ap.Labels["env"],
			Summary:      ap.Annotations["summary"],
			Description:  ap.Annotations["description"],
			Runbook:      ap.Annotations["runbook_url"],
			StartsAt:     ap.StartsAt,
			EndsAt:       ap.EndsAt,
			CreatedAt:    now,
			Labels:       labels,
			Annotations:  annotations,
			GeneratorURL: ap.GeneratorURL,
			Receiver:     payload.Receiver,
			GroupKey:     payload.GroupKey,
		}

		if err := db.Create(&alert).Error; err != nil {
			log.Printf("[Webhook] 存储告警失败 [%d/%d] %s: %v", i+1, len(payload.Alerts), alertname, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "store failed", "detail": err.Error()})
			return
		}
		log.Printf("[Webhook] 已存储 [%d/%d] %s | %s | %s | instance=%s",
			i+1, len(payload.Alerts), alertname, ap.Status, severity, ap.Labels["instance"])
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
