package model

import "time"

// InboundIngestLog stores a persistent audit trail of raw JSON telemetry payloads
// received by Cloud Mode ingest endpoints across all tenants (accessible strictly by SUPERADMIN).
type InboundIngestLog struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ReceivedAt         time.Time `gorm:"index;not null" json:"received_at"`
	TenantID           uint      `gorm:"index;default:0" json:"tenant_id"`
	TenantName         string    `gorm:"size:128" json:"tenant_name"`
	SiteID             uint      `gorm:"index;default:0" json:"site_id"`
	SiteCode           string    `gorm:"size:64;index" json:"site_code"`
	DeviceCode         string    `gorm:"size:64;index" json:"device_code"`
	Endpoint           string    `gorm:"size:128;index" json:"endpoint"`
	Source             string    `gorm:"size:32;index" json:"source"`          // EDGE_PUSH | MQTT
	LoggerType         string    `gorm:"size:32;index" json:"logger_type"`     // AQMS | SPARING | ONLIMO | WQMS
	DetectedFormat     string    `gorm:"size:48;index" json:"detected_format"` // STANDARD_V1 | CBI_LEGACY | KLHK_SPARING | etc.
	Status             string    `gorm:"size:24;index" json:"status"`          // ACCEPTED | REJECTED
	HTTPStatus         int       `gorm:"default:200" json:"http_status"`
	ClientIP           string    `gorm:"size:64" json:"client_ip"`
	KeyPrefix          string    `gorm:"size:32" json:"key_prefix"`
	RecordsProcessed   int       `gorm:"default:0" json:"records_processed"`
	ParametersIngested int       `gorm:"default:0" json:"parameters_ingested"`
	ParametersSkipped  int       `gorm:"default:0" json:"parameters_skipped"`
	PayloadBytes       int       `gorm:"default:0" json:"payload_bytes"`
	LatencyMs          int64     `gorm:"default:0" json:"latency_ms"`
	ErrorMessage       string    `gorm:"type:text" json:"error_message,omitempty"`
	RawPayload         string    `gorm:"type:longtext" json:"raw_payload"`
}

func (InboundIngestLog) TableName() string {
	return "inbound_ingest_logs"
}
