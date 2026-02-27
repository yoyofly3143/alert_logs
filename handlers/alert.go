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
	Alertname   string             `json:"alertname"`
	Status      models.AlertStatus `json:"status"`
	Severity    string             `json:"severity"`
	StartsAt    time.Time          `json:"starts_at"`
	EndsAt      *time.Time         `json:"ends_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	Labels      models.JSONMap     `json:"labels"`
	Annotations models.JSONMap     `json:"annotations"`
}

// GetAlerts returns paginated list of alerts
// Query params: page, page_size, status, severity, alertname, start_date, end_date
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
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if alertname := c.Query("alertname"); alertname != "" {
		query = query.Where("alertname LIKE ?", "%"+alertname+"%")
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

	// Get total count
	var total int64
	query.Count(&total)

	// Get alerts
	var alerts []models.Alert
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&alerts)
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
			Alertname:   alert.Alertname,
			Status:      alert.Status,
			Severity:    alert.Severity,
			StartsAt:    alert.StartsAt,
			EndsAt:      alert.EndsAt,
			CreatedAt:   alert.CreatedAt,
			Labels:      alert.Labels,
			Annotations: alert.Annotations,
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
	BySeverity   []SeverityCount  `json:"by_severity"`
	ByAlertname  []AlertnameCount `json:"by_alertname"`
	RecentFiring int64            `json:"recent_firing"`
	TodayAlerts  int64            `json:"today_alerts"`
}

// StatusCount represents count by status
type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// SeverityCount represents count by severity
type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

// AlertnameCount represents count by alertname
type AlertnameCount struct {
	Alertname string `json:"alertname"`
	Count     int64  `json:"count"`
}

// GetStats returns alert statistics
func (h *AlertHandler) GetStats(c *gin.Context) {
	db := database.GetDB()

	stats := AlertStats{}

	// Total count
	db.Model(&models.Alert{}).Count(&stats.Total)

	// By status
	var statusCounts []StatusCount
	db.Model(&models.Alert{}).Select("status as status, COUNT(*) as count").Group("status").Scan(&statusCounts)
	stats.ByStatus = statusCounts

	// By severity
	var severityCounts []SeverityCount
	db.Model(&models.Alert{}).Select("COALESCE(severity, 'unknown') as severity, COUNT(*) as count").Group("severity").Order("count DESC").Limit(10).Scan(&severityCounts)
	stats.BySeverity = severityCounts

	// By alertname (top 10)
	var alertnameCounts []AlertnameCount
	db.Model(&models.Alert{}).Select("alertname, COUNT(*) as count").Group("alertname").Order("count DESC").Limit(10).Scan(&alertnameCounts)
	stats.ByAlertname = alertnameCounts

	// Recent firing (last 24 hours)
	var recentFiring int64
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	db.Model(&models.Alert{}).Where("status = ? AND starts_at >= ?", models.StatusFiring, oneDayAgo).Count(&recentFiring)
	stats.RecentFiring = recentFiring

	// Today's alerts
	var todayAlerts int64
	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.Alert{}).Where("created_at >= ?", today).Count(&stats.TodayAlerts)
	stats.TodayAlerts = todayAlerts

	c.JSON(200, stats)
}

// GetSeverities returns all unique severities
func (h *AlertHandler) GetSeverities(c *gin.Context) {
	db := database.GetDB()

	var severities []string
	db.Model(&models.Alert{}).Distinct("severity").Pluck("severity", &severities)

	c.JSON(200, gin.H{"severities": severities})
}

// GetAlertNames returns all unique alert names
func (h *AlertHandler) GetAlertNames(c *gin.Context) {
	db := database.GetDB()

	var alertnames []string
	db.Model(&models.Alert{}).Distinct("alertname").Pluck("alertname", &alertnames)

	c.JSON(200, gin.H{"alertnames": alertnames})
}
