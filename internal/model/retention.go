package model

import (
	"time"
)

// RetentionCategory defines supported data domains for lifecycle management
type RetentionCategory string

const (
	CategoryRawTelemetry         RetentionCategory = "RAW_TELEMETRY"
	CategoryInternalAggregations RetentionCategory = "INTERNAL_AGGREGATIONS"
	CategoryCustomerAggregations RetentionCategory = "CUSTOMER_AGGREGATIONS"
	CategoryClearedAlarms        RetentionCategory = "CLEARED_ALARMS"
	CategorySystemLogs           RetentionCategory = "SYSTEM_LOGS"
	CategoryCommunicationLogs    RetentionCategory = "COMMUNICATION_LOGS"
	CategoryAuditTrails          RetentionCategory = "AUDIT_TRAILS"
	CategoryBackupArchives       RetentionCategory = "BACKUP_ARCHIVES"
)

// ValidCategories provides a list of all recognized retention categories
var ValidCategories = []RetentionCategory{
	CategoryRawTelemetry,
	CategoryInternalAggregations,
	CategoryCustomerAggregations,
	CategoryClearedAlarms,
	CategorySystemLogs,
	CategoryCommunicationLogs,
	CategoryAuditTrails,
	CategoryBackupArchives,
}

// RetentionPolicy defines independent rules for pruning old records in a specific category
type RetentionPolicy struct {
	ID                     string            `gorm:"primaryKey;size:64" json:"id"`
	Name                   string            `gorm:"size:128;not null" json:"name"`
	Category               RetentionCategory `gorm:"size:64;not null;uniqueIndex:idx_ret_cat" json:"category"`
	Description            string            `gorm:"size:255" json:"description"`
	Enabled                bool              `gorm:"default:false;index" json:"enabled"` // Must default to FALSE on new installations
	RetentionDays          int               `gorm:"not null;default:30" json:"retention_days"` // Days to keep before deletion
	MinimumAgeHours        int               `gorm:"not null;default:24" json:"minimum_age_hours"` // Safety buffer; records younger cannot be deleted
	ProtectedPeriodDays    int               `gorm:"not null;default:0" json:"protected_period_days"`
	RequireBackup          bool              `gorm:"not null" json:"require_backup"` // If true, target window must be covered by a verified backup
	RequireRollup          bool              `gorm:"default:false" json:"require_rollup"` // For raw telemetry: verify aggregations exist before deleting
	BatchSize              int               `gorm:"not null;default:500" json:"batch_size"` // Bounded deletion chunk size
	MaxDeletePerRun        int64             `gorm:"not null;default:50000" json:"max_delete_per_run"` // Ceiling per execution run (0 = unlimited bounded by batching)
	Priority               int               `gorm:"not null;default:10" json:"priority"` // Execution priority (lower = run earlier)
	ScheduleIntervalHours  int               `gorm:"not null;default:24" json:"schedule_interval_hours"` // Scheduled housekeeping interval
	ScheduleTime           string            `gorm:"size:16;default:'03:30'" json:"schedule_time"` // Time of day for housekeeping (HH:MM)

	// Runtime & Execution Tracking
	LastExecutionTime   *time.Time `json:"last_execution_time,omitempty"`
	LastExecutionResult string     `gorm:"size:32;default:'NONE'" json:"last_execution_result"` // SUCCESS, SKIPPED, BLOCKED, FAILED
	LastDeletedCount    int64      `gorm:"default:0" json:"last_deleted_count"`
	LastDurationMs      int64      `gorm:"default:0" json:"last_duration_ms"`
	NextScheduledTime   *time.Time `json:"next_scheduled_time,omitempty"`

	// Actor and Audit Metadata
	CreatedBy string    `gorm:"size:64;default:'system'" json:"created_by"`
	UpdatedBy string    `gorm:"size:64;default:'system'" json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RetentionExecutionType distinguishes simulation preview from physical deletion
type RetentionExecutionType string

const (
	ExecutionDryRun  RetentionExecutionType = "DRY_RUN"
	ExecutionExecute RetentionExecutionType = "EXECUTE"
)

// RetentionExecutionStatus tracks the state of an individual retention job run
type RetentionExecutionStatus string

const (
	RetentionStatusRunning   RetentionExecutionStatus = "RUNNING"
	RetentionStatusCompleted RetentionExecutionStatus = "COMPLETED"
	RetentionStatusBlocked   RetentionExecutionStatus = "BLOCKED"
	RetentionStatusFailed    RetentionExecutionStatus = "FAILED"
	RetentionStatusCancelled RetentionExecutionStatus = "CANCELLED"
)

// RetentionExecutionLog records auditable history for dry runs and destructive cleanups
type RetentionExecutionLog struct {
	ID                     uint64                   `gorm:"primaryKey;autoIncrement" json:"id"`
	PolicyID               string                   `gorm:"size:64;index;not null" json:"policy_id"`
	Category               RetentionCategory        `gorm:"size:64;index;not null" json:"category"`
	ExecutionType          RetentionExecutionType   `gorm:"size:16;index;not null" json:"execution_type"`
	TriggerType            string                   `gorm:"size:32;not null" json:"trigger_type"` // MANUAL, SCHEDULED
	Status                 RetentionExecutionStatus `gorm:"size:32;index;not null" json:"status"`
	CutoffTimestamp        time.Time                `gorm:"not null" json:"cutoff_timestamp"`
	CandidateCount         int64                    `gorm:"default:0" json:"candidate_count"`
	DeletedCount           int64                    `gorm:"default:0" json:"deleted_count"`
	BytesRecoveredEstimate int64                    `gorm:"default:0" json:"bytes_recovered_estimate"`
	BackupSafetyVerified   bool                     `gorm:"default:false" json:"backup_safety_verified"`
	BackupEvidence         string                   `gorm:"size:512" json:"backup_evidence,omitempty"`
	RollupSafetyVerified   bool                     `gorm:"default:false" json:"rollup_safety_verified"`
	RollupEvidence         string                   `gorm:"size:512" json:"rollup_evidence,omitempty"`
	BlockingReasons        string                   `gorm:"type:text" json:"blocking_reasons,omitempty"`
	ErrorMessage           string                   `gorm:"type:text" json:"error_message,omitempty"`
	DurationMs             int64                    `gorm:"default:0" json:"duration_ms"`
	InitiatedBy            string                   `gorm:"size:64;not null" json:"initiated_by"`
	CreatedAt              time.Time                `gorm:"index;not null" json:"created_at"`
}

// DryRunResult provides an uncommitted, read-only preview of candidate cleanup impact
type DryRunResult struct {
	PolicyID               string            `json:"policy_id"`
	PolicyName             string            `json:"policy_name"`
	Category               RetentionCategory `json:"category"`
	CutoffTimestamp        time.Time         `json:"cutoff_timestamp"`
	CandidateCount         int64             `json:"candidate_count"`
	ProtectedRecords       int64             `json:"protected_records"`
	EstimatedBytes         int64             `json:"estimated_bytes"`
	EstimatedMB            float64           `json:"estimated_mb"`
	BackupSafetyVerified   bool              `json:"backup_safety_verified"`
	BackupEvidence         string            `json:"backup_evidence"`
	RollupSafetyVerified   bool              `json:"rollup_safety_verified"`
	RollupEvidence         string            `json:"rollup_evidence"`
	CanExecute             bool              `json:"can_execute"`
	BlockingReasons        []string          `json:"blocking_reasons"`
	Dependencies           []string          `json:"dependencies"`
	ExecutedAt             time.Time         `json:"executed_at"`
}

// TableStorageInfo breaks down database storage footprint per table
type TableStorageInfo struct {
	TableName      string  `json:"table_name"`
	Category       string  `json:"category"`
	RowCount       int64   `json:"row_count"`
	DataSizeBytes  int64   `json:"data_size_bytes"`
	IndexSizeBytes int64   `json:"index_size_bytes"`
	TotalSizeBytes int64   `json:"total_size_bytes"`
	TotalSizeMB    float64 `json:"total_size_mb"`
	IsEstimated    bool    `json:"is_estimated"`
}

// StorageOverview provides comprehensive system-wide storage telemetry and disk capacity health
type StorageOverview struct {
	DatabaseType              string             `json:"database_type"`
	DatabaseName              string             `json:"database_name"`
	TotalDatabaseSizeBytes    int64              `json:"total_database_size_bytes"`
	TotalDatabaseSizeMB       float64            `json:"total_database_size_mb"`
	Tables                    []TableStorageInfo `json:"tables"`
	WALQueueSizeBytes         int64              `json:"wal_queue_size_bytes"`
	WALQueueSizeMB            float64            `json:"wal_queue_size_mb"`
	WALQueuePendingRecords    int64              `json:"wal_queue_pending_records"`
	BackupStorageSizeBytes    int64              `json:"backup_storage_size_bytes"`
	BackupStorageSizeMB       float64            `json:"backup_storage_size_mb"`
	BackupArchivesCount       int                `json:"backup_archives_count"`
	FilesystemTotalBytes      uint64             `json:"filesystem_total_bytes"`
	FilesystemFreeBytes       uint64             `json:"filesystem_free_bytes"`
	FilesystemUsedPercent     float64            `json:"filesystem_used_percent"`
	SharedFilesystemDetected  bool               `json:"shared_filesystem_detected"`
	CapacityStatus            string             `json:"capacity_status"` // HEALTHY, WARNING, CRITICAL
	WarningThresholdPercent   float64            `json:"warning_threshold_percent"`
	CriticalThresholdPercent  float64            `json:"critical_threshold_percent"`
	LastHousekeepingTime      *time.Time         `json:"last_housekeeping_time,omitempty"`
	LastHousekeepingStatus    string             `json:"last_housekeeping_status"`
	ActiveHousekeepingRunning bool               `json:"active_housekeeping_running"`
}
