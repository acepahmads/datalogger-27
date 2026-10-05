package model

import (
	"time"
)

type DestinationType string

const (
	DestREST      DestinationType = "REST_API"
	DestMQTT      DestinationType = "MQTT"
	DestTCP       DestinationType = "TCP"
	DestUDP       DestinationType = "UDP"
	DestWebSocket DestinationType = "WEBSOCKET"
	DestCloud     DestinationType = "CLOUD"
)

type OutputDestination struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	Name           string          `gorm:"size:128;not null" json:"name"`
	Type           DestinationType `gorm:"size:32;not null" json:"type"`
	Endpoint       string          `gorm:"size:255;not null" json:"endpoint"`
	Enabled        bool            `gorm:"default:false" json:"enabled"` // Local-first: default false unless configured!
	Format         string          `gorm:"size:32;default:'JSON'" json:"format"` // JSON, PROTOBUF, CSV
	BatchSize      int             `gorm:"default:100" json:"batch_size"`
	RetryLimit     int             `gorm:"default:5" json:"retry_limit"`
	QueueLimit     int             `gorm:"default:10000" json:"queue_limit"`
	Status         string          `gorm:"size:32;default:'IDLE'" json:"status"`
	LastDelivery   *time.Time      `json:"last_delivery"`
	SuccessDeliveries int64        `gorm:"default:0" json:"success_deliveries"`
	FailedDeliveries  int64        `gorm:"default:0" json:"failed_deliveries"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type DeliveryLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DestinationID uint      `gorm:"index;not null" json:"destination_id"`
	BatchCount    int       `json:"batch_count"`
	Status        string    `gorm:"size:32;not null" json:"status"` // SUCCESS, FAILED, RETRYING
	HTTPCode      int       `json:"http_code"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`
	LatencyMs     int       `json:"latency_ms"`
	DeliveredAt   time.Time `gorm:"index;not null" json:"delivered_at"`
}
