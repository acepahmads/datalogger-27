package model

import (
	"time"

	"gorm.io/gorm"
)

// TelemetryQuality defines industrial data quality flags
type TelemetryQuality string

const (
	QualityGood      TelemetryQuality = "GOOD"
	QualityBad       TelemetryQuality = "BAD"
	QualityUncertain TelemetryQuality = "UNCERTAIN"
	QualityUnknown   TelemetryQuality = "UNKNOWN"
)

// RawData represents persisted raw and normalized telemetry from industrial sensors (Section 4 & 6)
type RawData struct {
	ID          uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID    uint             `gorm:"index:idx_raw_dev_param_rec,priority:1;index:idx_raw_dev_rec,priority:1;not null" json:"device_id"`
	ParameterID uint             `gorm:"index:idx_raw_dev_param_rec,priority:2;index:idx_raw_param;not null" json:"parameter_id"`

	// Engineering Scaled Values
	Value        float64          `gorm:"type:double" json:"value"`
	ValueNumeric *float64         `gorm:"type:double" json:"value_numeric,omitempty"`
	FormulaValue *float64         `gorm:"type:double" json:"formula_value,omitempty"`
	IsHeldValue  bool             `gorm:"default:false;index" json:"is_held_value"`
	ValueText    string           `gorm:"size:255" json:"value_text,omitempty"`
	ValueBool    *bool            `json:"value_bool,omitempty"`
	RawValue     float64          `gorm:"type:double" json:"raw_value"`
	RawHex       string           `gorm:"size:255" json:"raw_hex,omitempty"`
	RawBytes     []byte           `gorm:"type:blob" json:"raw_bytes,omitempty"`

	// Industrial Quality & Protocol Source
	Quality      TelemetryQuality `gorm:"size:32;index;default:'GOOD'" json:"quality"`
	Source       string           `gorm:"size:32;default:'MODBUS_TCP'" json:"source"`
	Sequence     uint64           `gorm:"index" json:"sequence"`

	// Explicit Timestamps (Section 6)
	DeviceTimestamp *time.Time   `gorm:"index" json:"device_timestamp,omitempty"`
	ReceivedAt      time.Time    `gorm:"index:idx_raw_dev_param_rec,priority:3;index:idx_raw_dev_rec,priority:2;index:idx_raw_rec;not null" json:"received_at"`
	StoredAt        time.Time    `gorm:"not null" json:"stored_at"`
	Timestamp       time.Time    `gorm:"index" json:"timestamp"` // Backward compatibility alias

	CreatedAt       time.Time    `json:"created_at"`
}

func (r *RawData) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().UTC()
	if r.ReceivedAt.IsZero() {
		r.ReceivedAt = now
	}
	if r.StoredAt.IsZero() {
		r.StoredAt = now
	}
	if r.Timestamp.IsZero() {
		r.Timestamp = r.ReceivedAt
	}
	if r.Quality == "" {
		r.Quality = QualityGood
	}
	if r.ValueNumeric == nil {
		v := r.Value
		r.ValueNumeric = &v
	}
	return nil
}

type ProcessedData struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	DeviceID    uint      `gorm:"index;not null" json:"device_id"`
	ParameterID uint      `gorm:"index;not null" json:"parameter_id"`
	Timestamp   time.Time `gorm:"index;not null" json:"timestamp"`
	Value       float64   `json:"value"`
	ScaledValue float64   `json:"scaled_value"`
	IsValid     bool      `gorm:"default:true" json:"is_valid"`
	QualityCode int       `gorm:"default:192" json:"quality_code"` // 192 = Good in OPC
	CreatedAt   time.Time `json:"created_at"`
}

type AggregatedData struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	DeviceID    uint      `gorm:"index;not null" json:"device_id"`
	ParameterID uint      `gorm:"index;not null" json:"parameter_id"`
	Interval    string    `gorm:"size:32;not null" json:"interval"` // 1m, 5m, 1h, 1d
	BucketTime  time.Time `gorm:"index;not null" json:"bucket_time"`
	AvgValue    float64   `json:"avg_value"`
	MinValue    float64   `json:"min_value"`
	MaxValue    float64   `json:"max_value"`
	SampleCount int       `json:"sample_count"`
	CreatedAt   time.Time `json:"created_at"`
}

