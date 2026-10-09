package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"

	"github.com/shirou/gopsutil/v3/disk"
	"gorm.io/gorm"
)

var (
	ErrRetentionOpRunning   = errors.New("another retention or housekeeping operation is currently running")
	ErrPolicyNotFound       = errors.New("retention policy not found")
	ErrPolicyDisabled       = errors.New("retention policy is disabled; enable policy before execution")
	ErrInvalidPolicyConfig  = errors.New("invalid retention policy configuration")
	ErrRetentionSafetyBlock = errors.New("retention execution blocked by safety checks")
)

type RetentionService struct {
	db               *gorm.DB
	cfg              *config.Config
	retentionRepo    *repository.RetentionRepository
	backupRepo       *repository.BackupRepository
	systemRepo       *repository.SystemRepository
	telemetryService *TelemetryService
	backupService    *BackupService

	opMu      sync.Mutex
	isRunning bool
	activeID  string
}

func NewRetentionService(
	db *gorm.DB,
	cfg *config.Config,
	retentionRepo *repository.RetentionRepository,
	backupRepo *repository.BackupRepository,
	systemRepo *repository.SystemRepository,
	telemetryService *TelemetryService,
	backupService *BackupService,
) *RetentionService {
	if cfg == nil {
		cfg = config.Get()
	}

	s := &RetentionService{
		db:               db,
		cfg:              cfg,
		retentionRepo:    retentionRepo,
		backupRepo:       backupRepo,
		systemRepo:       systemRepo,
		telemetryService: telemetryService,
		backupService:    backupService,
	}

	// Auto-migrate tables and seed default policies (all default to disabled)
	if retentionRepo != nil {
		_ = retentionRepo.AutoMigrate()
		_ = retentionRepo.SeedDefaultPolicies()
	}

	return s
}

// ----------------- Policy Management -----------------

// GetAllPolicies returns all configured retention policies
func (s *RetentionService) GetAllPolicies() ([]model.RetentionPolicy, error) {
	if s.retentionRepo == nil {
		return []model.RetentionPolicy{}, nil
	}
	return s.retentionRepo.GetAllPolicies()
}

// GetPolicyByID returns a single policy
func (s *RetentionService) GetPolicyByID(id string) (*model.RetentionPolicy, error) {
	if s.retentionRepo == nil {
		return nil, ErrPolicyNotFound
	}
	return s.retentionRepo.GetPolicyByID(id)
}

// ValidatePolicy strictly verifies policy configuration before persistence
func (s *RetentionService) ValidatePolicy(policy *model.RetentionPolicy) error {
	if policy == nil {
		return errors.New("policy cannot be nil")
	}
	if strings.TrimSpace(policy.ID) == "" {
		return errors.New("policy ID is required")
	}
	if strings.TrimSpace(policy.Name) == "" {
		return errors.New("policy name is required")
	}

	// Category check
	validCategory := false
	for _, c := range model.ValidCategories {
		if policy.Category == c {
			validCategory = true
			break
		}
	}
	if !validCategory {
		return fmt.Errorf("unrecognized retention category: %s", policy.Category)
	}

	// Critical Rule: prevent 0 or negative days that would cause immediate data wipe
	if policy.RetentionDays < 1 {
		return errors.New("retention_days must be at least 1 day (accidental zero or negative retention prohibited)")
	}
	if policy.MinimumAgeHours < 1 {
		return errors.New("minimum_age_hours must be at least 1 hour")
	}
	if policy.BatchSize < 10 || policy.BatchSize > 5000 {
		return errors.New("batch_size must be between 10 and 5000 records")
	}
	if policy.ScheduleIntervalHours < 1 {
		return errors.New("schedule_interval_hours must be at least 1 hour")
	}

	return nil
}

// SavePolicy validates and updates an existing policy or creates a new one
func (s *RetentionService) SavePolicy(policy *model.RetentionPolicy, actor string) error {
	if err := s.ValidatePolicy(policy); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPolicyConfig, err)
	}

	if actor == "" {
		actor = "administrator"
	}

	existing, err := s.retentionRepo.GetPolicyByID(policy.ID)
	if err != nil {
		policy.CreatedBy = actor
		policy.UpdatedBy = actor
		policy.CreatedAt = time.Now()
		policy.UpdatedAt = time.Now()
		if err := s.retentionRepo.CreatePolicy(policy); err != nil {
			return err
		}
		s.logAudit("POLICY_CREATED", policy.ID, actor, fmt.Sprintf("Created retention policy '%s' for category '%s'", policy.Name, policy.Category))
		return nil
	}

	// Update existing
	existing.Name = policy.Name
	existing.Description = policy.Description
	existing.Enabled = policy.Enabled
	existing.RetentionDays = policy.RetentionDays
	existing.MinimumAgeHours = policy.MinimumAgeHours
	existing.ProtectedPeriodDays = policy.ProtectedPeriodDays
	existing.RequireBackup = policy.RequireBackup
	existing.RequireRollup = policy.RequireRollup
	existing.BatchSize = policy.BatchSize
	existing.MaxDeletePerRun = policy.MaxDeletePerRun
	existing.Priority = policy.Priority
	existing.ScheduleIntervalHours = policy.ScheduleIntervalHours
	existing.ScheduleTime = policy.ScheduleTime
	existing.UpdatedBy = actor
	existing.UpdatedAt = time.Now()

	if err := s.retentionRepo.UpdatePolicy(existing); err != nil {
		return err
	}

	s.logAudit("POLICY_UPDATED", policy.ID, actor, fmt.Sprintf("Updated retention policy '%s' (enabled=%t, retention_days=%d)", existing.Name, existing.Enabled, existing.RetentionDays))
	return nil
}

// TogglePolicy enables or disables an existing policy
func (s *RetentionService) TogglePolicy(id string, enabled bool, actor string) error {
	if s.retentionRepo == nil {
		return ErrPolicyNotFound
	}
	if actor == "" {
		actor = "administrator"
	}
	if err := s.retentionRepo.TogglePolicy(id, enabled, actor); err != nil {
		return err
	}
	s.logAudit("POLICY_TOGGLED", id, actor, fmt.Sprintf("Toggled policy %s enabled=%t", id, enabled))
	return nil
}

// ----------------- Dry-Run & Safety Check Engine -----------------

// CalculateCutoff determines the safe deletion threshold timestamp
func (s *RetentionService) CalculateCutoff(policy *model.RetentionPolicy) time.Time {
	now := time.Now().UTC()
	cutoff := now.AddDate(0, 0, -policy.RetentionDays)

	// If ProtectedPeriodDays is configured, add extra buffer
	if policy.ProtectedPeriodDays > 0 {
		protectedCutoff := now.AddDate(0, 0, -(policy.RetentionDays + policy.ProtectedPeriodDays))
		if protectedCutoff.Before(cutoff) {
			cutoff = protectedCutoff
		}
	}

	// Enforce MinimumAgeHours boundary: records younger than MinimumAgeHours can NEVER be pruned
	maxAllowedCutoff := now.Add(-time.Duration(policy.MinimumAgeHours) * time.Hour)
	if cutoff.After(maxAllowedCutoff) {
		cutoff = maxAllowedCutoff
	}

	return cutoff
}

// DryRun evaluates candidate records for deletion without modifying any data
func (s *RetentionService) DryRun(ctx context.Context, policyID string) (*model.DryRunResult, error) {
	policy, err := s.retentionRepo.GetPolicyByID(policyID)
	if err != nil {
		return nil, ErrPolicyNotFound
	}

	cutoff := s.CalculateCutoff(policy)
	result := &model.DryRunResult{
		PolicyID:        policy.ID,
		PolicyName:      policy.Name,
		Category:        policy.Category,
		CutoffTimestamp: cutoff,
		CanExecute:      true,
		ExecutedAt:      time.Now().UTC(),
		BlockingReasons: make([]string, 0),
		Dependencies:    make([]string, 0),
	}

	// 1. Candidate counting & dependency identification
	candidateCount, protectedCount, avgBytes, err := s.evaluateCandidateRecords(ctx, policy.Category, cutoff)
	if err != nil {
		return nil, fmt.Errorf("candidate evaluation failed: %w", err)
	}
	result.CandidateCount = candidateCount
	result.ProtectedRecords = protectedCount
	result.EstimatedBytes = candidateCount * avgBytes
	result.EstimatedMB = float64(result.EstimatedBytes) / (1024 * 1024)

	// 2. Backup safety check
	if policy.RequireBackup {
		result.Dependencies = append(result.Dependencies, "Phase 4.3 Backup & Restore Catalog")
		covered, evidence, err := s.verifyBackupCoverage(ctx, policy.Category, cutoff)
		result.BackupSafetyVerified = covered
		result.BackupEvidence = evidence
		if err != nil || !covered {
			result.CanExecute = false
			result.BlockingReasons = append(result.BlockingReasons, evidence)
		}
	} else {
		result.BackupSafetyVerified = true
		result.BackupEvidence = "Policy does not require pre-cleanup backup verification"
	}

	// 3. Rollup safety check
	if policy.RequireRollup && policy.Category == model.CategoryRawTelemetry {
		result.Dependencies = append(result.Dependencies, "Phase 3.3 Aggregation Results (INTERNAL_RAW / CUSTOMER_PROCESSED)")
		covered, evidence, err := s.verifyRollupCoverage(ctx, cutoff)
		result.RollupSafetyVerified = covered
		result.RollupEvidence = evidence
		if err != nil || !covered {
			result.CanExecute = false
			result.BlockingReasons = append(result.BlockingReasons, evidence)
		}
	} else {
		result.RollupSafetyVerified = true
		result.RollupEvidence = "Rollup verification not required for this category"
	}

	// 4. Record audit dry-run log
	_ = s.retentionRepo.CreateExecutionLog(&model.RetentionExecutionLog{
		PolicyID:               policy.ID,
		Category:               policy.Category,
		ExecutionType:          model.ExecutionDryRun,
		TriggerType:            "MANUAL",
		Status:                 model.RetentionStatusCompleted,
		CutoffTimestamp:        cutoff,
		CandidateCount:         candidateCount,
		BytesRecoveredEstimate: result.EstimatedBytes,
		BackupSafetyVerified:   result.BackupSafetyVerified,
		BackupEvidence:         result.BackupEvidence,
		RollupSafetyVerified:   result.RollupSafetyVerified,
		RollupEvidence:         result.RollupEvidence,
		BlockingReasons:        strings.Join(result.BlockingReasons, "; "),
		DurationMs:             time.Since(result.ExecutedAt).Milliseconds(),
		InitiatedBy:            "dry_run",
		CreatedAt:              time.Now().UTC(),
	})

	return result, nil
}

// ----------------- Execution Engine -----------------

// ExecutePolicy performs bounded, transactional deletion of expired records
func (s *RetentionService) ExecutePolicy(ctx context.Context, policyID string, initiatedBy string, forceOverride bool) (*model.RetentionExecutionLog, error) {
	policy, err := s.retentionRepo.GetPolicyByID(policyID)
	if err != nil {
		return nil, ErrPolicyNotFound
	}

	if !policy.Enabled && !forceOverride {
		return nil, ErrPolicyDisabled
	}

	s.opMu.Lock()
	if s.isRunning {
		s.opMu.Unlock()
		return nil, ErrRetentionOpRunning
	}
	s.isRunning = true
	s.activeID = policyID
	s.opMu.Unlock()

	defer func() {
		s.opMu.Lock()
		s.isRunning = false
		s.activeID = ""
		s.opMu.Unlock()
	}()

	startTime := time.Now()
	cutoff := s.CalculateCutoff(policy)

	execLog := &model.RetentionExecutionLog{
		PolicyID:        policy.ID,
		Category:        policy.Category,
		ExecutionType:   model.ExecutionExecute,
		TriggerType:     initiatedBy,
		Status:          model.RetentionStatusRunning,
		CutoffTimestamp: cutoff,
		InitiatedBy:     initiatedBy,
		CreatedAt:       startTime.UTC(),
	}
	_ = s.retentionRepo.CreateExecutionLog(execLog)

	// 1. Verify safety conditions
	var blockingReasons []string

	if policy.RequireBackup {
		covered, evidence, err := s.verifyBackupCoverage(ctx, policy.Category, cutoff)
		execLog.BackupSafetyVerified = covered
		execLog.BackupEvidence = evidence
		if err != nil || !covered {
			blockingReasons = append(blockingReasons, evidence)
		}
	} else {
		execLog.BackupSafetyVerified = true
		execLog.BackupEvidence = "Policy does not require backup check"
	}

	if policy.RequireRollup && policy.Category == model.CategoryRawTelemetry {
		covered, evidence, err := s.verifyRollupCoverage(ctx, cutoff)
		execLog.RollupSafetyVerified = covered
		execLog.RollupEvidence = evidence
		if err != nil || !covered {
			blockingReasons = append(blockingReasons, evidence)
		}
	} else {
		execLog.RollupSafetyVerified = true
		execLog.RollupEvidence = "Rollup check not required"
	}

	if len(blockingReasons) > 0 && !forceOverride {
		execLog.Status = model.RetentionStatusBlocked
		execLog.BlockingReasons = strings.Join(blockingReasons, "; ")
		execLog.DurationMs = time.Since(startTime).Milliseconds()
		_ = s.retentionRepo.UpdateExecutionLog(execLog)
		_ = s.retentionRepo.UpdateExecutionResult(policy.ID, "BLOCKED", 0, execLog.DurationMs)
		s.logAudit("RETENTION_BLOCKED", policy.ID, initiatedBy, fmt.Sprintf("Execution blocked by safety checks: %s", execLog.BlockingReasons))
		return execLog, fmt.Errorf("%w: %s", ErrRetentionSafetyBlock, execLog.BlockingReasons)
	}

	// 2. Execute bounded batch deletions
	deletedCount, err := s.performBoundedDeletion(ctx, policy.Category, cutoff, policy.BatchSize, policy.MaxDeletePerRun)
	execLog.DurationMs = time.Since(startTime).Milliseconds()
	execLog.DeletedCount = deletedCount

	if err != nil {
		execLog.Status = model.RetentionStatusFailed
		execLog.ErrorMessage = err.Error()
		_ = s.retentionRepo.UpdateExecutionLog(execLog)
		_ = s.retentionRepo.UpdateExecutionResult(policy.ID, "FAILED", deletedCount, execLog.DurationMs)
		s.logAudit("RETENTION_FAILED", policy.ID, initiatedBy, fmt.Sprintf("Retention execution failed: %v", err))
		return execLog, err
	}

	execLog.Status = model.RetentionStatusCompleted
	_ = s.retentionRepo.UpdateExecutionLog(execLog)
	_ = s.retentionRepo.UpdateExecutionResult(policy.ID, "SUCCESS", deletedCount, execLog.DurationMs)

	s.logAudit("RETENTION_EXECUTED", policy.ID, initiatedBy,
		fmt.Sprintf("Pruned %d records from %s (cutoff: %s, duration: %dms)",
			deletedCount, policy.Category, cutoff.Format(time.RFC3339), execLog.DurationMs))

	logger.Info("Retention policy %s completed: %d records deleted from %s in %dms",
		policy.ID, deletedCount, policy.Category, execLog.DurationMs)

	return execLog, nil
}

// ----------------- Candidate Evaluation & Bounded Deletion -----------------

func (s *RetentionService) evaluateCandidateRecords(ctx context.Context, category model.RetentionCategory, cutoff time.Time) (int64, int64, int64, error) {
	var candidateCount int64
	var protectedCount int64
	var avgBytes int64 = 128

	switch category {
	case model.CategoryRawTelemetry:
		avgBytes = 180
		err := s.db.WithContext(ctx).Model(&model.RawData{}).Where("received_at < ?", cutoff).Count(&candidateCount).Error
		return candidateCount, 0, avgBytes, err

	case model.CategoryInternalAggregations:
		avgBytes = 220
		err := s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Where("source_type = ? AND period_start < ?", model.SourceInternalRaw, cutoff).
			Count(&candidateCount).Error
		return candidateCount, 0, avgBytes, err

	case model.CategoryCustomerAggregations:
		avgBytes = 220
		err := s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Where("source_type = ? AND period_start < ?", model.SourceCustomerProcessed, cutoff).
			Count(&candidateCount).Error
		return candidateCount, 0, avgBytes, err

	case model.CategoryClearedAlarms:
		avgBytes = 250
		// Only count CLEARED alarms as candidates
		err := s.db.WithContext(ctx).Model(&model.Alarm{}).
			Where("status = ? AND (cleared_at < ? OR (cleared_at IS NULL AND triggered_at < ?))", model.AlarmCleared, cutoff, cutoff).
			Count(&candidateCount).Error
		if err != nil {
			return 0, 0, avgBytes, err
		}
		// Active & unacknowledged alarms are protected
		_ = s.db.WithContext(ctx).Model(&model.Alarm{}).
			Where("status != ?", model.AlarmCleared).
			Count(&protectedCount)
		return candidateCount, protectedCount, avgBytes, nil

	case model.CategorySystemLogs:
		avgBytes = 160
		err := s.db.WithContext(ctx).Model(&model.SystemLog{}).Where("created_at < ?", cutoff).Count(&candidateCount).Error
		return candidateCount, 0, avgBytes, err

	case model.CategoryCommunicationLogs:
		avgBytes = 350
		err := s.db.WithContext(ctx).Model(&model.CommunicationLog{}).Where("created_at < ?", cutoff).Count(&candidateCount).Error
		return candidateCount, 0, avgBytes, err

	case model.CategoryAuditTrails:
		avgBytes = 280
		err := s.db.WithContext(ctx).Model(&model.AuditTrail{}).Where("created_at < ?", cutoff).Count(&candidateCount).Error
		return candidateCount, 0, avgBytes, err

	case model.CategoryBackupArchives:
		avgBytes = 1024 * 1024
		var bErr error
		if s.backupRepo != nil {
			candidateCount, bErr = s.backupRepo.CountValid()
		}
		return candidateCount, 0, avgBytes, bErr

	default:
		return 0, 0, avgBytes, fmt.Errorf("unsupported category for evaluation: %s", category)
	}
}

func (s *RetentionService) performBoundedDeletion(
	ctx context.Context,
	category model.RetentionCategory,
	cutoff time.Time,
	batchSize int,
	maxDelete int64,
) (int64, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	var totalDeleted int64 = 0

	for {
		select {
		case <-ctx.Done():
			return totalDeleted, ctx.Err()
		default:
		}

		if maxDelete > 0 && totalDeleted >= maxDelete {
			break
		}

		limit := batchSize
		if maxDelete > 0 && int64(limit) > (maxDelete-totalDeleted) {
			limit = int(maxDelete - totalDeleted)
		}

		batchDeleted, err := s.deleteNextBatch(ctx, category, cutoff, limit)
		if err != nil {
			return totalDeleted, err
		}
		if batchDeleted == 0 {
			break // No more candidate records older than cutoff
		}
		totalDeleted += batchDeleted

		// Yield briefly to avoid saturating storage I/O on edge devices
		time.Sleep(5 * time.Millisecond)
	}

	return totalDeleted, nil
}

func (s *RetentionService) deleteNextBatch(ctx context.Context, category model.RetentionCategory, cutoff time.Time, limit int) (int64, error) {
	switch category {
	case model.CategoryRawTelemetry:
		var ids []uint64
		err := s.db.WithContext(ctx).Model(&model.RawData{}).
			Select("id").
			Where("received_at < ?", cutoff).
			Order("id ASC").
			Limit(limit).
			Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return 0, err
		}
		res := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.RawData{})
		return res.RowsAffected, res.Error

	case model.CategoryInternalAggregations:
		var ids []uint64
		err := s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Select("id").
			Where("source_type = ? AND period_start < ?", model.SourceInternalRaw, cutoff).
			Order("id ASC").
			Limit(limit).
			Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return 0, err
		}
		res := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.AggregationResult{})
		return res.RowsAffected, res.Error

	case model.CategoryCustomerAggregations:
		var ids []uint64
		err := s.db.WithContext(ctx).Model(&model.AggregationResult{}).
			Select("id").
			Where("source_type = ? AND period_start < ?", model.SourceCustomerProcessed, cutoff).
			Order("id ASC").
			Limit(limit).
			Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return 0, err
		}
		res := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.AggregationResult{})
		return res.RowsAffected, res.Error

	case model.CategoryClearedAlarms:
		var ids []uint
		// Critical rule: strictly delete ONLY CLEARED alarms
		err := s.db.WithContext(ctx).Model(&model.Alarm{}).
			Select("id").
			Where("status = ? AND (cleared_at < ? OR (cleared_at IS NULL AND triggered_at < ?))", model.AlarmCleared, cutoff, cutoff).
			Order("id ASC").
			Limit(limit).
			Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return 0, err
		}
		res := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Alarm{})
		return res.RowsAffected, res.Error

	case model.CategorySystemLogs:
		var ids []uint
		err := s.db.WithContext(ctx).Model(&model.SystemLog{}).
			Select("id").
			Where("created_at < ?", cutoff).
			Order("id ASC").
			Limit(limit).
			Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return 0, err
		}
		res := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.SystemLog{})
		return res.RowsAffected, res.Error

	case model.CategoryCommunicationLogs:
		var ids []uint
		err := s.db.WithContext(ctx).Model(&model.CommunicationLog{}).
			Select("id").
			Where("created_at < ?", cutoff).
			Order("id ASC").
			Limit(limit).
			Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return 0, err
		}
		res := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.CommunicationLog{})
		return res.RowsAffected, res.Error

	case model.CategoryAuditTrails:
		var ids []uint
		err := s.db.WithContext(ctx).Model(&model.AuditTrail{}).
			Select("id").
			Where("created_at < ?", cutoff).
			Order("id ASC").
			Limit(limit).
			Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return 0, err
		}
		res := s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.AuditTrail{})
		return res.RowsAffected, res.Error

	default:
		return 0, fmt.Errorf("unsupported category for deletion: %s", category)
	}
}

// ----------------- Safety Protection Checks -----------------

func (s *RetentionService) verifyBackupCoverage(ctx context.Context, category model.RetentionCategory, cutoff time.Time) (bool, string, error) {
	if s.backupRepo == nil {
		return false, "Backup subsystem not initialized; cannot verify backup coverage", nil
	}

	// 1. Check latest completed backup
	latestBackup, err := s.backupRepo.GetLatestCompleted()
	if err != nil || latestBackup == nil {
		return false, "No completed backup archive exists in the catalog; cleanup blocked", nil
	}

	// 2. Verify physical archive presence and validation status
	if latestBackup.ValidationStatus != model.ValidationStatusValid {
		return false, fmt.Sprintf("Latest backup %s has validation status '%s'; must be VALID", latestBackup.ID, latestBackup.ValidationStatus), nil
	}

	if latestBackup.FilePath == "" {
		return false, fmt.Sprintf("Latest backup %s has no storage file path recorded", latestBackup.ID), nil
	}
	if fi, statErr := os.Stat(latestBackup.FilePath); statErr != nil || fi.Size() == 0 {
		return false, fmt.Sprintf("Latest backup archive file %s is missing or empty on storage disk", latestBackup.Filename), nil
	}

	// 3. Verify coverage boundary:
	// If the latest backup snapshot was taken at or after cutoff, it covers the candidate data.
	// Allow a grace window of 1 hour for background snapshot processing.
	snapshotCoverageTime := latestBackup.CreatedAt.Add(1 * time.Hour)
	if snapshotCoverageTime.Before(cutoff) {
		return false, fmt.Sprintf("Latest verified backup %s was taken at %s, which is older than target cutoff %s (uncovered window exists)",
			latestBackup.ID, latestBackup.CreatedAt.Format(time.RFC3339), cutoff.Format(time.RFC3339)), nil
	}

	evidence := fmt.Sprintf("Covered by verified backup %s (%s, %d bytes, status=%s)",
		latestBackup.ID, latestBackup.Filename, latestBackup.SizeBytes, latestBackup.ValidationStatus)
	return true, evidence, nil
}

func (s *RetentionService) verifyRollupCoverage(ctx context.Context, cutoff time.Time) (bool, string, error) {
	// Verify whether raw_data before cutoff has been rolled up into aggregation_results
	var activeDefsCount int64
	s.db.WithContext(ctx).Model(&model.AggregationDefinition{}).Where("enabled = ?", true).Count(&activeDefsCount)
	if activeDefsCount == 0 {
		return true, "No active aggregation definitions configured; rollup check passed", nil
	}

	// Check if any active definition has a LastCalculatedAt older than cutoff
	var pendingDefs []model.AggregationDefinition
	err := s.db.WithContext(ctx).Model(&model.AggregationDefinition{}).
		Where("enabled = ? AND (last_calculated_at IS NULL OR last_calculated_at < ?)", true, cutoff).
		Find(&pendingDefs).Error
	if err != nil {
		return false, "Failed inspecting aggregation calculation status", err
	}

	if len(pendingDefs) > 0 {
		defNames := make([]string, 0, len(pendingDefs))
		for _, d := range pendingDefs {
			defNames = append(defNames, d.Name)
		}
		evidence := fmt.Sprintf("%d active aggregation definition(s) have not calculated rollups up to cutoff %s (e.g., %s)",
			len(pendingDefs), cutoff.Format(time.RFC3339), strings.Join(defNames[:min(3, len(defNames))], ", "))
		return false, evidence, nil
	}

	return true, "All active downsampled rollups are up to date through cutoff timestamp", nil
}

// ----------------- Storage Monitoring -----------------

// GetStorageOverview compiles exhaustive database, WAL, backup, and filesystem metrics
func (s *RetentionService) GetStorageOverview() (*model.StorageOverview, error) {
	cfg := s.cfg
	if cfg == nil {
		cfg = config.Get()
	}

	overview := &model.StorageOverview{
		DatabaseType:             cfg.DBType,
		DatabaseName:             cfg.DBName,
		Tables:                   make([]model.TableStorageInfo, 0),
		WarningThresholdPercent:  80.0,
		CriticalThresholdPercent: 90.0,
		CapacityStatus:           "HEALTHY",
	}

	// 1. Inspect Table breakdown
	var totalDBBytes int64 = 0

	monitoredTables := []struct {
		Name     string
		Category string
	}{
		{"raw_data", "Raw Telemetry"},
		{"aggregation_results", "Downsampled Aggregations"},
		{"aggregation_definitions", "Aggregation Definitions"},
		{"devices", "Device Configuration"},
		{"parameters", "Parameter Configuration"},
		{"alarms", "Alarms & Notifications"},
		{"system_logs", "Application Logs"},
		{"communication_logs", "Protocol Frame Logs"},
		{"audit_trails", "Security Audit Trails"},
		{"backup_records", "Backup Catalog"},
		{"retention_policies", "Retention Policies"},
		{"retention_execution_logs", "Retention Execution History"},
	}

	for _, mt := range monitoredTables {
		info := model.TableStorageInfo{
			TableName: mt.Name,
			Category:  mt.Category,
		}

		if s.db.Migrator().HasTable(mt.Name) {
			var count int64
			_ = s.db.Table(mt.Name).Count(&count)
			info.RowCount = count

			// Query specific data & index length in MariaDB if available
			if strings.EqualFold(cfg.DBType, "mariadb") || strings.EqualFold(cfg.DBType, "mysql") {
				type TableSize struct {
					DataLength  int64 `gorm:"column:data_length"`
					IndexLength int64 `gorm:"column:index_length"`
				}
				var ts TableSize
				query := "SELECT IFNULL(data_length, 0) as data_length, IFNULL(index_length, 0) as index_length FROM information_schema.tables WHERE table_schema = ? AND table_name = ?"
				if err := s.db.Raw(query, cfg.DBName, mt.Name).Scan(&ts).Error; err == nil && (ts.DataLength > 0 || ts.IndexLength > 0) {
					info.DataSizeBytes = ts.DataLength
					info.IndexSizeBytes = ts.IndexLength
					info.TotalSizeBytes = ts.DataLength + ts.IndexLength
					info.IsEstimated = false
				} else {
					// Fallback to estimated row size
					info.DataSizeBytes = count * 200
					info.IndexSizeBytes = count * 64
					info.TotalSizeBytes = info.DataSizeBytes + info.IndexSizeBytes
					info.IsEstimated = true
				}
			} else {
				// SQLite estimation
				info.DataSizeBytes = count * 200
				info.IndexSizeBytes = count * 64
				info.TotalSizeBytes = info.DataSizeBytes + info.IndexSizeBytes
				info.IsEstimated = true
			}
		}

		info.TotalSizeMB = float64(info.TotalSizeBytes) / (1024 * 1024)
		totalDBBytes += info.TotalSizeBytes
		overview.Tables = append(overview.Tables, info)
	}

	overview.TotalDatabaseSizeBytes = totalDBBytes
	overview.TotalDatabaseSizeMB = float64(totalDBBytes) / (1024 * 1024)

	// 2. Inspect WAL queue spool size
	if s.telemetryService != nil {
		m := s.telemetryService.GetMetrics()
		overview.WALQueueSizeBytes = m.QueueSpoolSizeBytes
		overview.WALQueueSizeMB = float64(m.QueueSpoolSizeBytes) / (1024 * 1024)
		overview.WALQueuePendingRecords = m.QueuePendingRecords
	}

	// 3. Inspect Backup archives storage
	if s.backupService != nil {
		bStats := s.backupService.GetStatus()
		if usedBytes, ok := bStats["storage_used_bytes"].(int64); ok {
			overview.BackupStorageSizeBytes = usedBytes
			overview.BackupStorageSizeMB = float64(usedBytes) / (1024 * 1024)
		}
		if totalBackups, ok := bStats["total_valid_backups"].(int64); ok {
			overview.BackupArchivesCount = int(totalBackups)
		}
	}

	// 4. Filesystem capacity measurement (avoiding double-counting shared disks)
	dataDir := cfg.DataDir
	if dataDir == "" {
		dataDir = "data"
	}
	absDataDir, _ := filepath.Abs(dataDir)
	diskUsage, err := disk.Usage(absDataDir)
	if err == nil && diskUsage != nil {
		overview.FilesystemTotalBytes = diskUsage.Total
		overview.FilesystemFreeBytes = diskUsage.Free
		overview.FilesystemUsedPercent = diskUsage.UsedPercent

		if overview.FilesystemUsedPercent >= overview.CriticalThresholdPercent {
			overview.CapacityStatus = "CRITICAL"
		} else if overview.FilesystemUsedPercent >= overview.WarningThresholdPercent {
			overview.CapacityStatus = "WARNING"
		} else {
			overview.CapacityStatus = "HEALTHY"
		}
	}

	// Check shared filesystem between backup dir and data dir
	backupDir := cfg.BackupDir
	if backupDir == "" {
		backupDir = filepath.Join("data", "backups")
	}
	absBackupDir, _ := filepath.Abs(backupDir)
	backupDiskUsage, errB := disk.Usage(absBackupDir)
	if err == nil && errB == nil && diskUsage != nil && backupDiskUsage != nil {
		if diskUsage.Total == backupDiskUsage.Total && diskUsage.Free == backupDiskUsage.Free {
			overview.SharedFilesystemDetected = true
		}
	}

	// 5. Last housekeeping execution status
	latestLog, err := s.retentionRepo.GetLatestExecutionLog()
	if err == nil && latestLog != nil {
		overview.LastHousekeepingTime = &latestLog.CreatedAt
		overview.LastHousekeepingStatus = string(latestLog.Status)
	} else {
		overview.LastHousekeepingStatus = "NONE"
	}

	s.opMu.Lock()
	overview.ActiveHousekeepingRunning = s.isRunning
	s.opMu.Unlock()

	return overview, nil
}

// GetExecutionLogs returns paginated history
func (s *RetentionService) GetExecutionLogs(page, pageSize int, policyID string) ([]model.RetentionExecutionLog, int64, error) {
	if s.retentionRepo == nil {
		return []model.RetentionExecutionLog{}, 0, nil
	}
	return s.retentionRepo.GetExecutionLogs(page, pageSize, policyID)
}

// ----------------- Internal Helper Methods -----------------

func (s *RetentionService) logAudit(action, resource, username, details string) {
	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  username,
			Action:    action,
			Resource:  resource,
			Details:   details,
			IPAddress: "127.0.0.1",
			UserAgent: "RetentionService",
			CreatedAt: time.Now(),
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
