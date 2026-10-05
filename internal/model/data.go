package model

import (
	"time"
)

type RawData struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	DeviceID    uint      `gorm:"index;not null" json:"device_id"`
	ParameterID uint      `gorm:"index;not null" json:"parameter_id"`
	Timestamp   time.Time `gorm:"index;not null" json:"timestamp"`
	RawBytes    []byte    `json:"raw_bytes,omitempty"`
	RawValue    float64   `json:"raw_value"`
	Quality     string    `gorm:"size:32;default:'GOOD'" json:"quality"` // GOOD, BAD, UNCERTAIN, SUSPECT
	CreatedAt   time.Time `json:"created_at"`
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
