package model

import "time"

// TelemetryInjectionBatch tracks a Superadmin-only historical/raw/processed telemetry injection batch.
// Stored in a dedicated table so normal telemetry, audit trails, and system logs never expose injection metadata
// to non-Superadmin roles (ADMIN, ENGINEER, OPERATOR).
type TelemetryInjectionBatch struct {
	ID                   uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchCode            string     `gorm:"size:64;uniqueIndex;not null" json:"batch_code"`
	DeviceID             uint       `gorm:"index;not null" json:"device_id"`
	DeviceCode           string     `gorm:"size:64" json:"device_code"`
	DeviceName           string     `gorm:"size:128" json:"device_name"`
	StartTime            time.Time  `gorm:"index;not null" json:"start_time"`
	EndTime              time.Time  `gorm:"index;not null" json:"end_time"`
	SampleStepSeconds    int        `gorm:"not null;default:30" json:"sample_step_seconds"`
	ConflictMode         string     `gorm:"size:32;not null;default:'FILL_MISSING_AND_NULL'" json:"conflict_mode"` // FILL_MISSING_AND_NULL, OVERWRITE_ALL
	RandomMode           string     `gorm:"size:32;not null;default:'SMOOTH_WALK'" json:"random_mode"`             // SMOOTH_WALK, UNIFORM
	ValueTarget          string     `gorm:"size:32;not null;default:'PROCESSED'" json:"value_target"`              // PROCESSED, RAW
	ParametersSummary    string     `gorm:"size:512" json:"parameters_summary"`
	ParametersConfigJSON string     `gorm:"type:text" json:"parameters_config_json,omitempty"`
	InjectedRawCount     int        `gorm:"default:0" json:"injected_raw_count"`
	SkippedCount         int        `gorm:"default:0" json:"skipped_count"`
	OverwrittenCount     int        `gorm:"default:0" json:"overwritten_count"`
	AggregatedBuckets    int        `gorm:"default:0" json:"aggregated_buckets"`
	RefreshedOutbound    int        `gorm:"default:0" json:"refreshed_outbound"`
	Status               string     `gorm:"size:24;index;default:'ACTIVE'" json:"status"` // ACTIVE, ROLLED_BACK
	CreatedBy            string     `gorm:"size:64" json:"created_by"`
	Note                 string     `gorm:"size:255" json:"note,omitempty"`
	RolledBackAt         *time.Time `json:"rolled_back_at,omitempty"`
	RolledBackBy         string     `gorm:"size:64" json:"rolled_back_by,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// TelemetryInjectionRecord links each injected RawData primary key to its TelemetryInjectionBatch
// for O(1) Superadmin badge annotation and lossless primary-key rollback without altering raw_data schema.
type TelemetryInjectionRecord struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchID     uint      `gorm:"index:idx_inj_rec_batch;not null" json:"batch_id"`
	BatchCode   string    `gorm:"size:64;index" json:"batch_code"`
	RawDataID   uint64    `gorm:"uniqueIndex:idx_inj_rec_raw_id;not null" json:"raw_data_id"`
	DeviceID    uint      `gorm:"index:idx_inj_rec_dev_param_time,priority:1;not null" json:"device_id"`
	ParameterID uint      `gorm:"index:idx_inj_rec_dev_param_time,priority:2;not null" json:"parameter_id"`
	ReceivedAt  time.Time `gorm:"index:idx_inj_rec_dev_param_time,priority:3;not null" json:"received_at"`
}

