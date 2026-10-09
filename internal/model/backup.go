package model

import (
	"os"
	"time"

	"gorm.io/gorm"
)

type BackupType string

const (
	BackupTypeManual    BackupType = "MANUAL"
	BackupTypeScheduled BackupType = "SCHEDULED"
	BackupTypeSafety    BackupType = "SAFETY"
)

type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "PENDING"
	BackupStatusInProgress BackupStatus = "IN_PROGRESS"
	BackupStatusCompleted  BackupStatus = "COMPLETED"
	BackupStatusFailed     BackupStatus = "FAILED"
	BackupStatusInvalid    BackupStatus = "INVALID"
)

type ValidationStatus string

const (
	ValidationStatusPending ValidationStatus = "PENDING"
	ValidationStatusValid   ValidationStatus = "VALID"
	ValidationStatusInvalid ValidationStatus = "INVALID"
)

// BackupRecord tracks catalog entries persisted in the database
type BackupRecord struct {
	ID                 string           `gorm:"primaryKey;size:64" json:"id"`
	Filename           string           `gorm:"size:255;not null" json:"filename"`
	FilePath           string           `gorm:"size:512;not null" json:"file_path"`
	Type               BackupType       `gorm:"size:32;not null" json:"type"`
	Status             BackupStatus     `gorm:"size:32;not null;index" json:"status"`
	ValidationStatus   ValidationStatus `gorm:"size:32;not null;index" json:"validation_status"`
	ConsistencyStatus  string           `gorm:"size:32;default:'CONSISTENT'" json:"consistency_status"`
	SizeBytes          int64            `json:"size_bytes"`
	SHA256Checksum     string           `gorm:"size:64" json:"sha256_checksum"`
	FormatVersion      string           `gorm:"size:32" json:"format_version"`
	AppVersion         string           `gorm:"size:32" json:"app_version"`
	DatabaseEngine     string           `gorm:"size:64" json:"database_engine"`
	DatabaseName       string           `gorm:"size:64" json:"database_name"`
	DumpMethod         string           `gorm:"size:64" json:"dump_method"`
	TableCount         int              `json:"table_count"`
	TotalRows          int64            `json:"total_rows"`
	WALSegmentsCount   int              `json:"wal_segments_count"`
	WALPendingRecords  int64            `json:"wal_pending_records"`
	SnapshotBoundaryTS *time.Time       `json:"snapshot_boundary_ts"`
	CreatedBy          string           `gorm:"size:64" json:"created_by"`
	DurationMs         int64            `json:"duration_ms"`
	ErrorMessage       string           `gorm:"type:text" json:"error_message,omitempty"`
	ValidationDetails  string           `gorm:"type:text" json:"validation_details,omitempty"`
	CreatedAt          time.Time        `gorm:"index;not null" json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`

	// Virtual fields for frontend API contract parity
	BackupID       string `gorm:"-" json:"backup_id"`
	TotalSizeBytes int64  `gorm:"-" json:"total_size_bytes"`
	BackupType     string `gorm:"-" json:"backup_type"`
	IsCompatible   bool   `gorm:"-" json:"is_compatible"`
}

// PopulateVirtualFields populates derived and compatibility fields for UI/API consumption
func (b *BackupRecord) PopulateVirtualFields() {
	b.BackupID = b.ID
	b.TotalSizeBytes = b.SizeBytes
	b.BackupType = string(b.Type)

	if b.ConsistencyStatus == "" {
		if b.Status == BackupStatusCompleted {
			b.ConsistencyStatus = "CONSISTENT"
		} else {
			b.ConsistencyStatus = "INCONSISTENT"
		}
	}

	// Verify physical presence on storage disk
	fileExists := false
	if b.FilePath != "" {
		if fi, err := os.Stat(b.FilePath); err == nil && fi.Size() > 0 {
			fileExists = true
			if b.SizeBytes == 0 {
				b.SizeBytes = fi.Size()
				b.TotalSizeBytes = fi.Size()
			}
		}
	}

	// Compatible only if status completed, validation valid, supported version, and file exists
	b.IsCompatible = (b.Status == BackupStatusCompleted &&
		b.ValidationStatus == ValidationStatusValid &&
		(b.FormatVersion == "1.0" || b.FormatVersion == "") &&
		fileExists)

	if !fileExists && b.Status == BackupStatusCompleted && b.ValidationStatus == ValidationStatusValid {
		b.ValidationStatus = ValidationStatusInvalid
		b.ValidationDetails = "Backup archive file is missing from storage disk"
		b.ConsistencyStatus = "INCONSISTENT"
		b.IsCompatible = false
	}
}

// AfterFind GORM hook automatically executes PopulateVirtualFields on database queries
func (b *BackupRecord) AfterFind(tx *gorm.DB) (err error) {
	b.PopulateVirtualFields()
	return nil
}

// BackupManifest represents the machine-readable manifest.json inside each backup artifact
type BackupManifest struct {
	BackupID         string               `json:"backup_id"`
	FormatVersion    string               `json:"format_version"`
	AppVersion       string               `json:"app_version"`
	CreatedAt        time.Time            `json:"created_at"`
	BackupType       BackupType           `json:"backup_type"`
	Database         DatabaseManifestInfo `json:"database"`
	WAL              WALManifestInfo      `json:"wal"`
	ConfigSummary    SafeConfigExport     `json:"config_summary"`
	Artifacts        []ArtifactInfo       `json:"artifacts"`
	TotalSizeBytes   int64                `json:"total_size_bytes"`
	ArchiveSHA256    string               `json:"archive_sha256"`
	Status           BackupStatus         `json:"status"`
	ValidationStatus ValidationStatus     `json:"validation_status"`
}

type DatabaseManifestInfo struct {
	Engine        string           `json:"engine"`
	Host          string           `json:"host"`
	DatabaseName  string           `json:"database_name"`
	DumpMethod    string           `json:"dump_method"` // "mariadb-dump", "mysqldump", "go-stream-sql"
	DumpFileName  string           `json:"dump_filename"`
	DumpSHA256    string           `json:"dump_sha256"`
	DumpSizeBytes int64            `json:"dump_size_bytes"`
	TableCount    int              `json:"table_count"`
	TotalRows     int64            `json:"total_rows"`
	TableRows     map[string]int64 `json:"table_rows,omitempty"`
	Artifact      *ArtifactInfo    `json:"artifact,omitempty"`
}

type WALManifestInfo struct {
	PendingRecords    int64          `json:"pending_records"`
	SpoolSizeBytes    int64          `json:"spool_size_bytes"`
	CheckpointSegment uint32         `json:"checkpoint_segment"`
	CheckpointOffset  int64          `json:"checkpoint_offset"`
	ActiveSegment     uint32         `json:"active_segment"`
	ActiveOffset      int64          `json:"active_offset"`
	SegmentFiles      []ArtifactInfo `json:"segment_files"`
	Segments          []ArtifactInfo `json:"segments,omitempty"`
}

type SafeConfigExport struct {
	AppName               string `json:"app_name"`
	Environment           string `json:"environment"`
	Port                  string `json:"port"`
	DBType                string `json:"db_type"`
	DBHost                string `json:"db_host"`
	DBPort                string `json:"db_port"`
	DBName                string `json:"db_name"`
	QueueEnabled          bool   `json:"queue_enabled"`
	QueueSyncMode         string `json:"queue_sync_mode"`
	QueueMaxSizeBytes     int64  `json:"queue_max_size_bytes"`
	BackupDir             string `json:"backup_dir"`
	BackupScheduleEnabled bool   `json:"backup_schedule_enabled"`
}

type ArtifactInfo struct {
	Path        string `json:"path"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	CRC32       uint32 `json:"crc32,omitempty"`
	Description string `json:"description"`
}

type BackupJobStatus string

const (
	JobStatusQueued    BackupJobStatus = "QUEUED"
	JobStatusRunning   BackupJobStatus = "RUNNING"
	JobStatusCompleted BackupJobStatus = "COMPLETED"
	JobStatusFailed    BackupJobStatus = "FAILED"
)

type BackupJobType string

const (
	JobTypeBackup   BackupJobType = "BACKUP"
	JobTypeRestore  BackupJobType = "RESTORE"
	JobTypeValidate BackupJobType = "VALIDATE"
)

type BackupJob struct {
	ID          string          `json:"id"`
	Type        BackupJobType   `json:"type"`
	BackupID    string          `json:"backup_id,omitempty"`
	Status      BackupJobStatus `json:"status"`
	Progress    float64         `json:"progress"` // 0 - 100
	Stage       string          `json:"stage"`
	InitiatedBy string          `json:"initiated_by"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     *time.Time      `json:"end_time,omitempty"`
	Error       string          `json:"error,omitempty"`
}

type RestorePreview struct {
	BackupID             string           `json:"backup_id"`
	CreatedAt            time.Time        `json:"created_at"`
	BackupType           BackupType       `json:"backup_type"`
	SizeBytes            int64            `json:"size_bytes"`
	DatabaseEngine       string           `json:"database_engine"`
	DatabaseName         string           `json:"database_name"`
	DumpMethod           string           `json:"dump_method"`
	TableCount           int              `json:"table_count"`
	TotalRows            int64            `json:"total_rows"`
	TableRows            map[string]int64 `json:"table_rows"`
	WALPendingRecords    int64            `json:"wal_pending_records"`
	WALSegmentsCount     int              `json:"wal_segments_count"`
	CurrentActiveTables  int              `json:"current_active_tables"`
	CurrentActiveRows    int64            `json:"current_active_rows"`
	EstimatedDurationSec int              `json:"estimated_duration_sec"`
	TargetDatabase       string           `json:"target_database,omitempty"`
	Warnings             []string         `json:"warnings"`
	CanRestore           bool             `json:"can_restore"`
}

type BackupScheduleConfig struct {
	Enabled       bool       `json:"enabled"`
	IntervalHours int        `json:"interval_hours"`
	TimeOfDay     string     `json:"time_of_day"`
	Compression   string     `json:"compression"`
	KeepMaxCount  int        `json:"keep_max_count"`
	NextRunTime   *time.Time `json:"next_run_time"`
	LastRunTime   *time.Time `json:"last_run_time"`
	LastStatus    string     `json:"last_status"`
}
