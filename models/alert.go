package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type AlertStatus string

const (
	StatusFiring   AlertStatus = "firing"
	StatusResolved AlertStatus = "resolved"
)

// Alert represents a single alert from AlertManager
type Alert struct {
	ID           uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Fingerprint  string      `gorm:"size:64;not null;uniqueIndex" json:"fingerprint"`
	Alertname    string      `gorm:"size:255;not null;index" json:"alertname"`
	Status       AlertStatus `gorm:"type:enum('firing','resolved');not null;index" json:"status"`
	Severity     string      `gorm:"size:50;index" json:"severity"`
	StartsAt     time.Time   `gorm:"not null;index" json:"starts_at"`
	EndsAt       *time.Time  `json:"ends_at"`
	CreatedAt    time.Time   `gorm:"autoCreateTime" json:"created_at"`
	Labels       JSONMap     `gorm:"type:json;not null" json:"labels"`
	Annotations  JSONMap     `gorm:"type:json" json:"annotations"`
	GeneratorURL string      `gorm:"size:500" json:"generator_url"`
	Receiver     string      `gorm:"size:255" json:"receiver"`
	GroupKey     string      `gorm:"size:255;index" json:"group_key"`
}

func (Alert) TableName() string {
	return "alerts"
}

// JSONMap is a map[string]interface{} that implements GORM's Scanner/Valuer
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// WebhookPayload represents the AlertManager webhook payload
type WebhookPayload struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	Alerts            []AlertPayload    `json:"alerts"`
}

// AlertPayload represents a single alert in the webhook payload
type AlertPayload struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       *time.Time        `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}
