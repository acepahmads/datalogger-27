package repository

import (
	"errors"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

type RetentionRepository struct {
	db *gorm.DB
}

func NewRetentionRepository(db *gorm.DB) *RetentionRepository {
	return &RetentionRepository{db: db}
}

// AutoMigrate synchronizes schema for retention policies and execution logs
func (r *RetentionRepository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&model.RetentionPolicy{},
		&model.RetentionExecutionLog{},
	)
}

// SeedDefaultPolicies initializes default baseline policies in a DISABLED state
func (r *RetentionRepository) SeedDefaultPolicies() error {
	defaults := []model.RetentionPolicy{
		{
			ID:                    "pol_raw_telemetry",
			Name:                  "Raw Telemetry Retention",
			Category:              model.CategoryRawTelemetry,
			Description:           "Retain granular high-frequency raw sensor data; prune records rolled up into aggregations or backed up",
			Enabled:               false, // Crucial: default to disabled
			RetentionDays:         30,
			MinimumAgeHours:       24,
			ProtectedPeriodDays:   7,
			RequireBackup:         true,
			RequireRollup:         true,
			BatchSize:             500,
			MaxDeletePerRun:       50000,
			Priority:              10,
			ScheduleIntervalHours: 24,
			ScheduleTime:          "03:00",
			CreatedBy:             "system",
		},
		{
			ID:                    "pol_internal_agg",
			Name:                  "Internal Raw Aggregations Retention",
			Category:              model.CategoryInternalAggregations,
			Description:           "Retain intermediate engineering rollups; prune old internal buckets once historical trends are established",
			Enabled:               false,
			RetentionDays:         60,
			MinimumAgeHours:       48,
			ProtectedPeriodDays:   14,
			RequireBackup:         true,
			RequireRollup:         false,
			BatchSize:             500,
			MaxDeletePerRun:       20000,
			Priority:              20,
			ScheduleIntervalHours: 24,
			ScheduleTime:          "03:15",
			CreatedBy:             "system",
		},
		{
			ID:                    "pol_customer_agg",
			Name:                  "Customer Processed Aggregations Retention",
			Category:              model.CategoryCustomerAggregations,
			Description:           "Long-term customer reporting rollups; strictly separated from internal engineering aggregations",
			Enabled:               false,
			RetentionDays:         365,
			MinimumAgeHours:       168,
			ProtectedPeriodDays:   30,
			RequireBackup:         true,
			RequireRollup:         false,
			BatchSize:             500,
			MaxDeletePerRun:       10000,
			Priority:              30,
			ScheduleIntervalHours: 24,
			ScheduleTime:          "03:30",
			CreatedBy:             "system",
		},
		{
			ID:                    "pol_cleared_alarms",
			Name:                  "Cleared Alarms Retention",
			Category:              model.CategoryClearedAlarms,
			Description:           "Prune historical cleared alarms; active and unacknowledged alarms are permanently protected",
			Enabled:               false,
			RetentionDays:         90,
			MinimumAgeHours:       24,
			ProtectedPeriodDays:   7,
			RequireBackup:         false,
			RequireRollup:         false,
			BatchSize:             250,
			MaxDeletePerRun:       5000,
			Priority:              40,
			ScheduleIntervalHours: 24,
			ScheduleTime:          "03:45",
			CreatedBy:             "system",
		},
		{
			ID:                    "pol_comm_logs",
			Name:                  "Communication Frame Logs Retention",
			Category:              model.CategoryCommunicationLogs,
			Description:           "Prune high-volume raw Modbus serial and TCP frame logs used for diagnostics",
			Enabled:               false,
			RetentionDays:         14,
			MinimumAgeHours:       12,
			ProtectedPeriodDays:   3,
			RequireBackup:         false,
			RequireRollup:         false,
			BatchSize:             1000,
			MaxDeletePerRun:       100000,
			Priority:              5, // High priority to recover storage quickly
			ScheduleIntervalHours: 12,
			ScheduleTime:          "04:00",
			CreatedBy:             "system",
		},
		{
			ID:                    "pol_system_logs",
			Name:                  "System Diagnostics Logs Retention",
			Category:              model.CategorySystemLogs,
			Description:           "Prune general application debug and informational diagnostic log entries",
			Enabled:               false,
			RetentionDays:         30,
			MinimumAgeHours:       24,
			ProtectedPeriodDays:   7,
			RequireBackup:         false,
			RequireRollup:         false,
			BatchSize:             500,
			MaxDeletePerRun:       20000,
			Priority:              15,
			ScheduleIntervalHours: 24,
			ScheduleTime:          "04:15",
			CreatedBy:             "system",
		},
		{
			ID:                    "pol_audit_trails",
			Name:                  "Security Audit Trails Retention",
			Category:              model.CategoryAuditTrails,
			Description:           "Retain administrative and compliance audit history with extended retention",
			Enabled:               false,
			RetentionDays:         180,
			MinimumAgeHours:       720,
			ProtectedPeriodDays:   90,
			RequireBackup:         true,
			RequireRollup:         false,
			BatchSize:             250,
			MaxDeletePerRun:       5000,
			Priority:              50,
			ScheduleIntervalHours: 24,
			ScheduleTime:          "04:30",
			CreatedBy:             "system",
		},
	}

	for _, p := range defaults {
		var existing model.RetentionPolicy
		if err := r.db.Where("category = ? OR id = ?", p.Category, p.ID).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				_ = r.db.Create(&p).Error
			}
		}
	}
	return nil
}

// GetAllPolicies returns all configured retention policies ordered by priority
func (r *RetentionRepository) GetAllPolicies() ([]model.RetentionPolicy, error) {
	var policies []model.RetentionPolicy
	err := r.db.Order("priority ASC, id ASC").Find(&policies).Error
	return policies, err
}

// GetPolicyByID retrieves a policy by its unique identifier
func (r *RetentionRepository) GetPolicyByID(id string) (*model.RetentionPolicy, error) {
	var policy model.RetentionPolicy
	err := r.db.Where("id = ?", id).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// GetPolicyByCategory retrieves a policy by data domain category
func (r *RetentionRepository) GetPolicyByCategory(category model.RetentionCategory) (*model.RetentionPolicy, error) {
	var policy model.RetentionPolicy
	err := r.db.Where("category = ?", category).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// CreatePolicy inserts a new retention policy
func (r *RetentionRepository) CreatePolicy(policy *model.RetentionPolicy) error {
	return r.db.Create(policy).Error
}

// UpdatePolicy updates editable policy configuration
func (r *RetentionRepository) UpdatePolicy(policy *model.RetentionPolicy) error {
	policy.UpdatedAt = time.Now()
	return r.db.Save(policy).Error
}

// TogglePolicy enables or disables a policy
func (r *RetentionRepository) TogglePolicy(id string, enabled bool, updatedBy string) error {
	return r.db.Model(&model.RetentionPolicy{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enabled":    enabled,
		"updated_by": updatedBy,
		"updated_at": time.Now(),
	}).Error
}

// UpdateExecutionResult records the outcome of a dry-run or physical cleanup run
func (r *RetentionRepository) UpdateExecutionResult(id string, result string, deletedCount int64, durationMs int64) error {
	now := time.Now()
	return r.db.Model(&model.RetentionPolicy{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_execution_time":   &now,
		"last_execution_result": result,
		"last_deleted_count":    deletedCount,
		"last_duration_ms":      durationMs,
		"updated_at":            now,
	}).Error
}

// CreateExecutionLog inserts a historical execution record
func (r *RetentionRepository) CreateExecutionLog(log *model.RetentionExecutionLog) error {
	log.CreatedAt = time.Now()
	return r.db.Create(log).Error
}

// UpdateExecutionLog updates an ongoing or finished execution log
func (r *RetentionRepository) UpdateExecutionLog(log *model.RetentionExecutionLog) error {
	return r.db.Save(log).Error
}

// GetExecutionLogs returns paginated execution history with optional policy filter
func (r *RetentionRepository) GetExecutionLogs(page, pageSize int, policyID string) ([]model.RetentionExecutionLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := r.db.Model(&model.RetentionExecutionLog{})
	if policyID != "" {
		query = query.Where("policy_id = ?", policyID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.RetentionExecutionLog
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// GetLatestExecutionLog returns the most recent execution record
func (r *RetentionRepository) GetLatestExecutionLog() (*model.RetentionExecutionLog, error) {
	var log model.RetentionExecutionLog
	err := r.db.Order("id DESC").First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}
