package model

import (
	"time"
)

type AuditTrail struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `gorm:"index" json:"user_id"`
	Username  string    `gorm:"size:64;not null" json:"username"`
	Action    string    `gorm:"size:128;not null" json:"action"`
	Resource  string    `gorm:"size:128;not null" json:"resource"`
	Details   string    `gorm:"type:text" json:"details"`
	IPAddress string    `gorm:"size:64" json:"ip_address"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	CreatedAt time.Time `gorm:"index;not null" json:"created_at"`
}

type SystemLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Level     string    `gorm:"size:16;index;not null" json:"level"` // DEBUG, INFO, WARN, ERROR
	Component string    `gorm:"size:64;index;not null" json:"component"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Details   string    `gorm:"type:text" json:"details"`
	CreatedAt time.Time `gorm:"index;not null" json:"created_at"`
}

type CommunicationLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DeviceID     uint      `gorm:"index;not null" json:"device_id"`
	Protocol     string    `gorm:"size:32;not null" json:"protocol"`
	Direction    string    `gorm:"size:16;not null" json:"direction"` // TX, RX
	PayloadHex   string    `gorm:"type:text" json:"payload_hex"`
	Status       string    `gorm:"size:32;not null" json:"status"` // SUCCESS, TIMEOUT, ERROR
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	LatencyMs    int       `json:"latency_ms"`
	CreatedAt    time.Time `gorm:"index;not null" json:"created_at"`
}
