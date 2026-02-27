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

// Alert 严格按照用户要求的 10 个字段建立的表结构
type Alert struct {
	ID          uint64      `gorm:"primaryKey;autoIncrement"                          json:"id"`           // 1. 主键 ID
	Fingerprint string      `gorm:"size:64;not null;index:idx_fingerprint_starts,unique"            json:"fingerprint"`  // 2. 告警指纹
	Status      AlertStatus `gorm:"type:varchar(20);not null;index"                   json:"status"`       // 3. 状态
	AlertName   string      `gorm:"column:alert_name;size:128;not null;index"         json:"alert_name"`   // 4. 告警名称
	Instance    string      `gorm:"size:128;index"                                    json:"instance"`     // 5. 实例
	StartsAt    time.Time   `gorm:"type:datetime(3);not null;index:idx_fingerprint_starts,unique" json:"starts_at"`    // 6. 开始时间
	EndsAt      *time.Time  `gorm:"type:datetime(3)"                                              json:"ends_at"`      // 7. 结束时间
	Labels      JSONMap     `gorm:"type:json;not null"                                json:"labels"`       // 8. 存储所有标签
	Annotations JSONMap     `gorm:"type:json"                                         json:"annotations"`  // 9. 存储描述详情
	RawContent  JSONMap     `gorm:"column:raw_content;type:json"                       json:"raw_content"`  // 10. 原始 JSON 备份
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
