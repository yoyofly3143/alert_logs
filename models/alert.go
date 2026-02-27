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

// Alert 表示从 AlertManager 接收到的一条告警记录（每次push均插入新行）
type Alert struct {
	ID          uint64      `gorm:"primaryKey;autoIncrement"         json:"id"`
	Fingerprint string      `gorm:"size:64;not null;index"           json:"fingerprint"`
	Alertname   string      `gorm:"size:255;not null;index"          json:"alertname"`
	Status      AlertStatus `gorm:"type:varchar(20);not null;index"  json:"status"`
	Severity    string      `gorm:"size:50;index"                    json:"severity"`

	// 常用 label 字段（冗余存储，便于查询/排序）
	Instance string `gorm:"size:255;index"  json:"instance"`
	Job      string `gorm:"size:255;index"  json:"job"`
	Cluster  string `gorm:"size:255;index"  json:"cluster"`
	Env      string `gorm:"size:100;index"  json:"env"`

	// 常用 annotation 字段
	Summary     string `gorm:"size:1000"  json:"summary"`
	Description string `gorm:"type:text"  json:"description"`
	Runbook     string `gorm:"size:500"   json:"runbook"`

	StartsAt time.Time  `gorm:"not null;index"  json:"starts_at"`
	EndsAt   *time.Time `                        json:"ends_at"`

	CreatedAt time.Time `gorm:"autoCreateTime"  json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"  json:"updated_at"`

	// 完整的 JSON 字段（保留所有原始数据）
	Labels      JSONMap `gorm:"type:json;not null"  json:"labels"`
	Annotations JSONMap `gorm:"type:json"           json:"annotations"`

	GeneratorURL string `gorm:"size:1000"  json:"generator_url"`
	Receiver     string `gorm:"size:255"   json:"receiver"`
	GroupKey     string `gorm:"size:500;index"  json:"group_key"`
}

func (Alert) TableName() string {
	return "alerts"
}

// JSONMap 实现 GORM Scanner/Valuer 接口，用于存储 JSON 数据
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
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for JSONMap")
	}
	return json.Unmarshal(bytes, j)
}

// WebhookPayload 表示 AlertManager 发出的 Webhook 数据结构
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

// AlertPayload 表示 Webhook 中单条告警的原始结构
type AlertPayload struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       *time.Time        `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}
