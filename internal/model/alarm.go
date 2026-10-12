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
	Notes          string        `gorm:"type:text" json:"notes,omitempty"`
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
	ChannelWebhook  NotificationChannelType = "WEBHOOK"
)

type NotificationStatus string

const (
	NotificationPending   NotificationStatus = "PENDING"
	NotificationSending   NotificationStatus = "SENDING"
	NotificationSent      NotificationStatus = "SENT"
	NotificationFailed    NotificationStatus = "FAILED"
	NotificationRetrying  NotificationStatus = "RETRYING"
	NotificationCancelled NotificationStatus = "CANCELLED"
)

type NotificationChannel struct {
	ID               uint                    `gorm:"primaryKey" json:"id"`
	Name             string                  `gorm:"size:128;not null" json:"name"`
	Type             NotificationChannelType `gorm:"size:32;not null" json:"type"`
	Enabled          bool                    `gorm:"default:true" json:"enabled"`
	ConfigJSON       string                  `gorm:"type:text" json:"config_json,omitempty"` // sensitive config (stored securely)
	MinSeverity      AlarmSeverity           `gorm:"size:32;default:'WARNING'" json:"min_severity"`
	RateLimitPerMin  int                     `gorm:"default:10" json:"rate_limit_per_min"`
	MaxRetries       int                     `gorm:"default:3" json:"max_retries"`
	RetryIntervalSec int                     `gorm:"default:30" json:"retry_interval_sec"`
	QuietHoursStart  string                  `gorm:"size:8" json:"quiet_hours_start,omitempty"` // e.g. "22:00"
	QuietHoursEnd    string                  `gorm:"size:8" json:"quiet_hours_end,omitempty"`   // e.g. "06:00"
	Timezone         string                  `gorm:"size:64;default:'UTC'" json:"timezone"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

type Notification struct {
	ID               uint               `gorm:"primaryKey" json:"id"`
	AlarmID          uint               `gorm:"index;not null" json:"alarm_id"`
	Alarm            *Alarm             `gorm:"foreignKey:AlarmID" json:"alarm,omitempty"`
	ChannelID        *uint              `gorm:"index" json:"channel_id,omitempty"`
	Channel          string             `gorm:"size:32;not null" json:"channel"`
	Recipient        string             `gorm:"size:128;not null" json:"recipient"`
	Subject          string             `gorm:"size:255" json:"subject,omitempty"`
	Message          string             `gorm:"type:text;not null" json:"message"`
	Status           NotificationStatus `gorm:"size:32;default:'PENDING';index" json:"status"`
	Attempts         int                `gorm:"default:0" json:"attempts"`
	MaxAttempts      int                `gorm:"default:3" json:"max_attempts"`
	NextRetryAt      *time.Time         `gorm:"index" json:"next_retry_at,omitempty"`
	ErrorMsg         string             `gorm:"type:text" json:"error_msg,omitempty"`
	ProviderResponse string             `gorm:"type:text" json:"provider_response,omitempty"`
	IdempotencyKey   string             `gorm:"size:128;index" json:"idempotency_key,omitempty"`
	SentAt           *time.Time         `json:"sent_at,omitempty"`
	CreatedAt        time.Time          `gorm:"index" json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}
