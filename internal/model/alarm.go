package model

import (
	"time"
)

type AlarmSeverity string

const (
	SeverityInfo     AlarmSeverity = "INFO"
	SeverityWarning  AlarmSeverity = "WARNING"
	SeverityCritical AlarmSeverity = "CRITICAL"
)

type AlarmStatus string

const (
	AlarmActive   AlarmStatus = "ACTIVE"
	AlarmAcked    AlarmStatus = "ACKNOWLEDGED"
	AlarmCleared  AlarmStatus = "CLEARED"
)

type Alarm struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	DeviceID       *uint         `gorm:"index" json:"device_id"`
	Device         *Device       `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	ParameterID    *uint         `gorm:"index" json:"parameter_id"`
	Parameter      *Parameter    `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`
	AlarmCode      string        `gorm:"size:64;not null" json:"alarm_code"`
	Severity       AlarmSeverity `gorm:"size:32;default:'WARNING'" json:"severity"`
	Message        string        `gorm:"size:255;not null" json:"message"`
	TriggerValue   *float64      `json:"trigger_value"`
	ThresholdValue *float64      `json:"threshold_value"`
	Status         AlarmStatus   `gorm:"size:32;default:'ACTIVE'" json:"status"`
	TriggeredAt    time.Time     `gorm:"not null" json:"triggered_at"`
	AcknowledgedAt *time.Time    `json:"acknowledged_at"`
	AcknowledgedBy string        `gorm:"size:64" json:"acknowledged_by"`
	ClearedAt      *time.Time    `json:"cleared_at"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type NotificationChannelType string

const (
	ChannelWeb      NotificationChannelType = "WEB"
	ChannelEmail    NotificationChannelType = "EMAIL"
	ChannelSMS      NotificationChannelType = "SMS"
	ChannelTelegram NotificationChannelType = "TELEGRAM"
	ChannelWhatsApp NotificationChannelType = "WHATSAPP"
	ChannelPush     NotificationChannelType = "PUSH"
)

type NotificationChannel struct {
	ID          uint                    `gorm:"primaryKey" json:"id"`
	Name        string                  `gorm:"size:128;not null" json:"name"`
	Type        NotificationChannelType `gorm:"size:32;not null" json:"type"`
	Enabled     bool                    `gorm:"default:true" json:"enabled"`
	ConfigJSON  string                  `gorm:"type:text" json:"config_json"` // bot token, chat ID, smtp credentials etc.
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AlarmID   uint      `gorm:"index;not null" json:"alarm_id"`
	Alarm     *Alarm    `gorm:"foreignKey:AlarmID" json:"alarm,omitempty"`
	Channel   string    `gorm:"size:32;not null" json:"channel"`
	Recipient string    `gorm:"size:128;not null" json:"recipient"`
	Status    string    `gorm:"size:32;default:'PENDING'" json:"status"` // PENDING, SENT, FAILED
	ErrorMsg  string    `gorm:"type:text" json:"error_msg"`
	SentAt    *time.Time `json:"sent_at"`
	CreatedAt time.Time `json:"created_at"`
}
