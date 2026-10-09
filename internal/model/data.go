package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TelemetryQuality defines industrial data quality flags
type TelemetryQuality string

const (
	QualityGood      TelemetryQuality = "GOOD"
	QualityBad       TelemetryQuality = "BAD"
	QualityUncertain TelemetryQuality = "UNCERTAIN"
	QualityStale     TelemetryQuality = "STALE"
	QualityUnknown   TelemetryQuality = "UNKNOWN"
)

// QualityReason defines standardized machine-readable diagnostic reasons
type QualityReason string

const (
	ReasonNone                  QualityReason = "NONE"
	ReasonOutOfHardRange        QualityReason = "OUT_OF_HARD_RANGE"
	ReasonOutOfWarningRange     QualityReason = "OUT_OF_WARNING_RANGE"
	ReasonNullValue             QualityReason = "NULL_VALUE"
	ReasonNaNValue              QualityReason = "NAN_VALUE"
	ReasonInfiniteValue         QualityReason = "INFINITE_VALUE"
	ReasonInvalidTimestamp      QualityReason = "INVALID_TIMESTAMP"
	ReasonFutureTimestamp       QualityReason = "FUTURE_TIMESTAMP"
	ReasonExcessiveOldTimestamp QualityReason = "EXCESSIVELY_OLD_TIMESTAMP"
	ReasonStaleData             QualityReason = "STALE_DATA"
	ReasonSpikeDetected         QualityReason = "SPIKE_DETECTED"
	ReasonDuplicateData         QualityReason = "DUPLICATE_DATA"
	ReasonDecodingError         QualityReason = "DECODING_ERROR"
	ReasonCommunicationError    QualityReason = "COMMUNICATION_ERROR"
	ReasonHoldAnomalyGrace      QualityReason = "HOLD_ANOMALY_GRACE"
)

// QualityFlag constants for multi-flag tagging
const (
	FlagRangeViolation = "range_violation"
	FlagRangeWarning   = "range_warning"
	FlagSpike          = "spike"
	FlagStale          = "stale"
	FlagTimestampWarn  = "timestamp_warning"
	FlagDuplicate      = "duplicate"
	FlagDecodeError    = "decode_error"
	FlagCommError      = "comm_error"
	FlagHeldAnomaly    = "held_anomaly"
)

// RawData represents persisted raw and normalized telemetry from industrial sensors (Section 4 & 6)
type RawData struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	RecordUUID  string `gorm:"size:64;index:idx_raw_record_uuid;uniqueIndex:idx_raw_record_uuid_uniq" json:"record_uuid,omitempty"`
	DeviceID    uint   `gorm:"index:idx_raw_dev_param_rec,priority:1;index:idx_raw_dev_rec,priority:1;not null" json:"device_id"`
	ParameterID uint   `gorm:"index:idx_raw_dev_param_rec,priority:2;index:idx_raw_param;not null" json:"parameter_id"`

	// Engineering Scaled & Processed Values
	Value          float64  `gorm:"type:double" json:"value"`
	ProcessedValue float64  `gorm:"type:double" json:"processed_value"`
	ValueNumeric   *float64 `gorm:"type:double" json:"value_numeric,omitempty"`
	FormulaValue   *float64 `gorm:"type:double" json:"formula_value,omitempty"`
	IsHeldValue    bool     `gorm:"default:false;index" json:"is_held_value"`
	ValueText      string   `gorm:"size:255" json:"value_text,omitempty"`
	ValueBool      *bool    `json:"value_bool,omitempty"`
	RawValue       float64  `gorm:"type:double" json:"raw_value"`
	RawHex         string   `gorm:"size:255" json:"raw_hex,omitempty"`
	RawBytes       []byte   `gorm:"type:blob" json:"raw_bytes,omitempty"`

	// Industrial Quality & Protocol Source (Phase 3.2)
	Quality       TelemetryQuality `gorm:"size:32;index;default:'GOOD'" json:"quality"`
	QualityReason QualityReason    `gorm:"size:64;index;default:'NONE'" json:"quality_reason"`
	QualityFlags  string           `gorm:"size:255;default:''" json:"quality_flags,omitempty"`
	Source        string           `gorm:"size:32;default:'MODBUS_TCP'" json:"source"`
	Sequence      uint64           `gorm:"index" json:"sequence"`

	// Explicit Timestamps (Section 6 & Phase 3.2)
	DeviceTimestamp *time.Time `gorm:"index" json:"device_timestamp,omitempty"`
	ReceivedAt      time.Time  `gorm:"index:idx_raw_dev_param_rec,priority:3;index:idx_raw_dev_rec,priority:2;index:idx_raw_rec;not null" json:"received_at"`
	ProcessedAt     *time.Time `gorm:"index" json:"processed_at,omitempty"`
	StoredAt        time.Time  `gorm:"not null" json:"stored_at"`
	Timestamp       time.Time  `gorm:"index" json:"timestamp"` // Backward compatibility alias

	CreatedAt time.Time `json:"created_at"`
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
	if r.RecordUUID == "" {
		r.RecordUUID = fmt.Sprintf("REC-%d-%d-%d-%d", r.DeviceID, r.ParameterID, r.Sequence, r.ReceivedAt.UnixNano())
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

