package handlers

import (
	"alert-webhook/database"
	"alert-webhook/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AlertHandler handles alert query requests
type AlertHandler struct{}

func NewAlertHandler() *AlertHandler {
	return &AlertHandler{}
}

// AlertListResponse represents the response for alert list
type AlertListResponse struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Alerts   []AlertListItem `json:"alerts"`
}

// AlertListItem represents a single alert in list response
type AlertListItem struct {
	ID          uint64             `json:"id"`
	Fingerprint string             `json:"fingerprint"`
	AlertName   string             `json:"alert_name"`
	Status      models.AlertStatus `json:"status"`
	Instance    string             `json:"instance"`
	StartsAt    time.Time          `json:"starts_at"`
	EndsAt      *time.Time         `json:"ends_at,omitempty"`
	Labels      models.JSONMap     `json:"labels"`
	Annotations models.JSONMap     `json:"annotations"`
	RawContent  models.JSONMap     `json:"raw_content,omitempty"`
}

// GetAlerts returns paginated list of alerts
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	db := database.GetDB()

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Build query
	query := db.Model(&models.Alert{})

	// Filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if alertname := c.Query("alert_name"); alertname != "" {
		query = query.Where("alert_name LIKE ?", "%"+alertname+"%")
	}
	if instance := c.Query("instance"); instance != "" {
		query = query.Where("instance LIKE ?", "%"+instance+"%")
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("starts_at >= ?", t)
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			t = t.Add(24 * time.Hour)
			query = query.Where("starts_at <= ?", t)
		}
	}
	if quality := c.Query("quality"); quality != "" {
		query = query.Where("labels->>'$.quality' = ?", quality)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get alerts
	var alerts []models.Alert
	result := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&alerts)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Failed to query alerts", "detail": result.Error.Error()})
		return
	}

	// Convert to response
	items := make([]AlertListItem, len(alerts))
	for i, alert := range alerts {
		items[i] = AlertListItem{
			ID:          alert.ID,
			Fingerprint: alert.Fingerprint,
			AlertName:   alert.AlertName,
			Status:      alert.Status,
			Instance:    alert.Instance,
			StartsAt:    alert.StartsAt,
			EndsAt:      alert.EndsAt,
			Labels:      alert.Labels,
			Annotations: alert.Annotations,
			RawContent:  alert.RawContent,
		}
	}

	c.JSON(200, AlertListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Alerts:   items,
	})
}

// GetAlertByID returns a single alert by ID
func (h *AlertHandler) GetAlertByID(c *gin.Context) {
	db := database.GetDB()
	id := c.Param("id")

	var alert models.Alert
	result := db.First(&alert, id)
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(404, gin.H{"error": "Alert not found"})
		return
	}
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Failed to query alert", "detail": result.Error.Error()})
		return
	}

	c.JSON(200, alert)
}

// AlertStats represents alert statistics
type AlertStats struct {
	Total        int64            `json:"total"`
	ByStatus     []StatusCount    `json:"by_status"`
	ByAlertName  []AlertNameCount `json:"by_alert_name"`
	RecentFiring int64            `json:"recent_firing"`
	TodayAlerts  int64            `json:"today_alerts"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type AlertNameCount struct {
	AlertName string `json:"alert_name"`
	Count     int64  `json:"count"`
}

// GetStats returns alert statistics
func (h *AlertHandler) GetStats(c *gin.Context) {
	db := database.GetDB()
	stats := AlertStats{}

	db.Model(&models.Alert{}).Count(&stats.Total)

	var statusCounts []StatusCount
	db.Model(&models.Alert{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts)
	stats.ByStatus = statusCounts

	var alertnameCounts []AlertNameCount
	db.Model(&models.Alert{}).Select("alert_name, COUNT(*) as count").Group("alert_name").Order("count DESC").Limit(10).Scan(&alertnameCounts)
	stats.ByAlertName = alertnameCounts

	var recentFiring int64
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	db.Model(&models.Alert{}).Where("status = ? AND starts_at >= ?", models.StatusFiring, oneDayAgo).Count(&recentFiring)
	stats.RecentFiring = recentFiring

	var todayAlerts int64
	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.Alert{}).Where("starts_at >= ?", today).Count(&todayAlerts)
	stats.TodayAlerts = todayAlerts

	c.JSON(200, stats)
}

// GetSeverities - Deprecated but keeping for UI compatibility if needed, returns from labels
func (h *AlertHandler) GetSeverities(c *gin.Context) {
	c.JSON(200, gin.H{"severities": []string{"critical", "warning", "info"}})
}

// GetAlertNames returns all unique alert names
func (h *AlertHandler) GetAlertNames(c *gin.Context) {
	db := database.GetDB()
	var alertnames []string
	db.Model(&models.Alert{}).Distinct("alert_name").Pluck("alert_name", &alertnames)
	c.JSON(200, gin.H{"alert_names": alertnames})
}
