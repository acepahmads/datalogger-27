package model

import (
	"time"

	"gorm.io/gorm"
)

// AggregationSourceType distinguishes between internal raw sensor data and customer-facing processed data
type AggregationSourceType string

const (
	SourceInternalRaw        AggregationSourceType = "INTERNAL_RAW"
	SourceCustomerProcessed  AggregationSourceType = "CUSTOMER_PROCESSED"
)

// AggregationFunction defines mathematical aggregation operations
type AggregationFunction string

const (
	FunctionAvg   AggregationFunction = "AVG"
	FunctionMin   AggregationFunction = "MIN"
	FunctionMax   AggregationFunction = "MAX"
	FunctionSum   AggregationFunction = "SUM"
	FunctionCount AggregationFunction = "COUNT"
	FunctionFirst AggregationFunction = "FIRST"
	FunctionLast  AggregationFunction = "LAST"
)

// AggregationDefinition represents configuration for periodic telemetry rollups (Phase 3.3)
type AggregationDefinition struct {
	ID                 uint                  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name               string                `gorm:"size:128;not null" json:"name"`
	Code               string                `gorm:"size:64;not null;index:idx_agg_def_dev_code,unique" json:"code"`
	SourceType         AggregationSourceType `gorm:"size:32;not null;default:'CUSTOMER_PROCESSED';index" json:"source_type"`
	DeviceID           uint                  `gorm:"not null;index;index:idx_agg_def_dev_code,unique" json:"device_id"`
	ParameterID        uint                  `gorm:"not null;index" json:"parameter_id"`
	Function           AggregationFunction   `gorm:"size:16;not null;default:'AVG'" json:"function"`
	IntervalSeconds    int                   `gorm:"not null;default:300;index" json:"interval_seconds"` // 120, 300, 1800, 3600, etc.
	Timezone           string                `gorm:"size:64;default:'UTC'" json:"timezone"`
	Enabled            bool                  `gorm:"default:true;index" json:"enabled"`
	QualityPolicy      string                `gorm:"size:32;default:'STANDARD'" json:"quality_policy"` // STANDARD, STRICT, PERMISSIVE
	IdentifierFormat   string                `gorm:"size:32;default:'YYYYMMDDHHmmss'" json:"identifier_format"`
	DestinationType    string                `gorm:"size:64;default:'INTERNAL'" json:"destination_type"`
	GracePeriodSeconds int                   `gorm:"default:120" json:"grace_period_seconds"` // Grace period for late data
	LastCalculatedAt   *time.Time            `json:"last_calculated_at,omitempty"`
	StartAt            *time.Time            `json:"start_at,omitempty"`
	EndAt              *time.Time            `json:"end_at,omitempty"`

	// Relationships
	Device    *Device    `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Parameter *Parameter `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// AggregationResult stores persisted rollups with deterministic period identifiers (Section 3.3.15)
type AggregationResult struct {
	ID                      uint64                `gorm:"primaryKey;autoIncrement" json:"id"`
	AggregationDefinitionID uint                  `gorm:"not null;index:idx_agg_res_def_start,unique;index:idx_agg_res_def" json:"aggregation_definition_id"`
	DeviceID                uint                  `gorm:"not null;index:idx_agg_res_dev_param" json:"device_id"`
	ParameterID             uint                  `gorm:"not null;index:idx_agg_res_dev_param;index:idx_agg_res_param" json:"parameter_id"`
	SourceType              AggregationSourceType `gorm:"size:32;not null;index" json:"source_type"`

	// Relationships
	Device    *Device    `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Parameter *Parameter `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`

	// Deterministic Time Bucket & Period Identifier (YYYYMMDDHHmmss)
	PeriodStart time.Time `gorm:"not null;index:idx_agg_res_def_start,unique;index:idx_agg_res_start" json:"period_start"`
	PeriodEnd   time.Time `gorm:"not null;index" json:"period_end"`
	Identifier  string    `gorm:"size:32;not null;index:idx_agg_res_identifier" json:"identifier"`

	Function        AggregationFunction `gorm:"size:16;not null" json:"function"`
	IntervalSeconds int                 `gorm:"not null;index" json:"interval_seconds"`

	// Numerical Aggregation Outputs
	Value      *float64 `gorm:"type:double" json:"value"` // Primary aggregate output (nil if no valid samples)
	MinValue   *float64 `gorm:"type:double" json:"min_value,omitempty"`
	MaxValue   *float64 `gorm:"type:double" json:"max_value,omitempty"`
	AvgValue   *float64 `gorm:"type:double" json:"avg_value,omitempty"`
	SumValue   *float64 `gorm:"type:double" json:"sum_value,omitempty"`
	FirstValue *float64 `gorm:"type:double" json:"first_value,omitempty"`
	LastValue  *float64 `gorm:"type:double" json:"last_value,omitempty"`

	// Sample & Quality Accounting (Section 3.3.10)
	SampleCount    int              `gorm:"not null;default:0" json:"sample_count"`
	ValidCount     int              `gorm:"not null;default:0" json:"valid_count"`
	GoodCount      int              `gorm:"not null;default:0" json:"good_count"`
	UncertainCount int              `gorm:"not null;default:0" json:"uncertain_count"`
	BadCount       int              `gorm:"not null;default:0" json:"bad_count"`
	StaleCount     int              `gorm:"not null;default:0" json:"stale_count"`
	Quality        TelemetryQuality `gorm:"size:32;not null;default:'GOOD';index" json:"quality"`
	QualityReason  QualityReason    `gorm:"size:64;not null;default:'NONE'" json:"quality_reason"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
