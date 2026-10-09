package service

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/model"
	"datalogger/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestRetentionEnvironment(t *testing.T) (*gorm.DB, *config.Config, *RetentionService, *repository.RetentionRepository, *repository.BackupRepository, *repository.SystemRepository, func()) {
	baseTmp, err := os.MkdirTemp("", "phase44_retention_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	backupDir := filepath.Join(baseTmp, "backups")
	dataDir := filepath.Join(baseTmp, "data")
	_ = os.MkdirAll(backupDir, 0755)
	_ = os.MkdirAll(dataDir, 0755)

	cfg := &config.Config{
		AppName:   "Datalogger Test",
		DBType:    "sqlite",
		DBName:    "datalogger_test",
		DataDir:   dataDir,
		BackupDir: backupDir,
	}

	dsn := fmt.Sprintf("file:mem_ret_db_%d?mode=memory&cache=shared", time.Now().UnixNano()+int64(rand.Intn(1000)))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test sqlite db: %v", err)
	}

	// AutoMigrate all relevant models
	err = db.AutoMigrate(
		&model.RetentionPolicy{},
		&model.RetentionExecutionLog{},
		&model.RawData{},
		&model.AggregationResult{},
		&model.AggregationDefinition{},
		&model.Alarm{},
		&model.SystemLog{},
		&model.CommunicationLog{},
		&model.AuditTrail{},
		&model.BackupRecord{},
		&model.SystemHealth{},
	)
	if err != nil {
		t.Fatalf("failed migrating test models: %v", err)
	}

	retentionRepo := repository.NewRetentionRepository(db)
	backupRepo := repository.NewBackupRepository(db)
	systemRepo := repository.NewSystemRepository(db)

	retentionService := NewRetentionService(db, cfg, retentionRepo, backupRepo, systemRepo, nil, nil)

	cleanup := func() {
		_ = os.RemoveAll(baseTmp)
	}

	return db, cfg, retentionService, retentionRepo, backupRepo, systemRepo, cleanup
}

func TestRetentionPolicyValidation(t *testing.T) {
	_, _, svc, _, _, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	// 1. Valid policy passes
	validPolicy := &model.RetentionPolicy{
		ID:                    "test_valid",
		Name:                  "Valid Policy",
		Category:              model.CategoryRawTelemetry,
		RetentionDays:         30,
		MinimumAgeHours:       24,
		BatchSize:             500,
		ScheduleIntervalHours: 24,
	}
	if err := svc.ValidatePolicy(validPolicy); err != nil {
		t.Errorf("expected valid policy to pass, got: %v", err)
	}

	// 2. Prohibit zero or negative retention days (Critical Safety Rule)
	zeroDays := &model.RetentionPolicy{
		ID:                    "test_zero",
		Name:                  "Zero Days Policy",
		Category:              model.CategoryRawTelemetry,
		RetentionDays:         0,
		MinimumAgeHours:       24,
		BatchSize:             500,
		ScheduleIntervalHours: 24,
	}
	if err := svc.ValidatePolicy(zeroDays); err == nil {
		t.Errorf("expected error for 0 retention days, got nil")
	}

	// 3. Prohibit invalid category
	badCat := &model.RetentionPolicy{
		ID:                    "test_bad_cat",
		Name:                  "Bad Category",
		Category:              "UNKNOWN_CATEGORY",
		RetentionDays:         30,
		MinimumAgeHours:       24,
		BatchSize:             500,
		ScheduleIntervalHours: 24,
	}
	if err := svc.ValidatePolicy(badCat); err == nil {
		t.Errorf("expected error for unknown category, got nil")
	}

	// 4. Batch size bounds
	badBatch := &model.RetentionPolicy{
		ID:                    "test_bad_batch",
		Name:                  "Bad Batch",
		Category:              model.CategoryRawTelemetry,
		RetentionDays:         30,
		MinimumAgeHours:       24,
		BatchSize:             0,
		ScheduleIntervalHours: 24,
	}
	if err := svc.ValidatePolicy(badBatch); err == nil {
		t.Errorf("expected error for batch size 0, got nil")
	}
}

func TestDefaultPoliciesSeededAndDisabled(t *testing.T) {
	_, _, svc, _, _, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	policies, err := svc.GetAllPolicies()
	if err != nil {
		t.Fatalf("failed retrieving policies: %v", err)
	}
	if len(policies) == 0 {
		t.Fatalf("expected default policies to be seeded, got 0")
	}

	// Critical Safety Rule: ALL default policies must start DISABLED
	for _, p := range policies {
		if p.Enabled {
			t.Errorf("policy %s must default to enabled=false, got true", p.ID)
		}
	}

	// Attempting to execute a disabled policy must return ErrPolicyDisabled
	_, err = svc.ExecutePolicy(context.Background(), policies[0].ID, "test", false)
	if err != ErrPolicyDisabled {
		t.Errorf("expected ErrPolicyDisabled for disabled policy, got: %v", err)
	}
}

func TestDryRunProducesNoSideEffects(t *testing.T) {
	db, _, svc, _, _, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	oldTime := time.Now().AddDate(0, 0, -60)
	recentTime := time.Now().AddDate(0, 0, -5)

	// Seed 10 old records and 5 recent records with unique UUIDs
	for i := 1; i <= 10; i++ {
		db.Create(&model.RawData{
			RecordUUID: fmt.Sprintf("old-raw-%d", i),
			DeviceID:   1,
			ReceivedAt: oldTime,
		})
	}
	for i := 1; i <= 5; i++ {
		db.Create(&model.RawData{
			RecordUUID: fmt.Sprintf("recent-raw-%d", i),
			DeviceID:   1,
			ReceivedAt: recentTime,
		})
	}

	var countBefore int64
	db.Model(&model.RawData{}).Count(&countBefore)
	if countBefore != 15 {
		t.Fatalf("expected 15 records before dry run, got %d", countBefore)
	}

	// Run dry run
	result, err := svc.DryRun(context.Background(), "pol_raw_telemetry")
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	if result.CandidateCount != 10 {
		t.Errorf("expected 10 candidate records in dry run, got %d", result.CandidateCount)
	}

	// Verify ZERO records were deleted (No side-effects rule)
	var countAfter int64
	db.Model(&model.RawData{}).Count(&countAfter)
	if countAfter != 15 {
		t.Errorf("dry run modified data! Expected 15 records, got %d", countAfter)
	}
}

func TestCutoffCalculationBoundaries(t *testing.T) {
	_, _, svc, _, _, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	now := time.Now().UTC()
	policy := &model.RetentionPolicy{
		RetentionDays:       30,
		MinimumAgeHours:     48,
		ProtectedPeriodDays: 5,
	}

	cutoff := svc.CalculateCutoff(policy)
	expectedMaxCutoff := now.Add(-48 * time.Hour)

	if cutoff.After(expectedMaxCutoff) {
		t.Errorf("cutoff %v should not exceed MinimumAgeHours boundary %v", cutoff, expectedMaxCutoff)
	}

	// With ProtectedPeriodDays=5, cutoff should be at least 35 days in the past
	expectedMinPast := now.AddDate(0, 0, -35)
	if cutoff.After(expectedMinPast.Add(1 * time.Hour)) {
		t.Errorf("cutoff %v should account for ProtectedPeriodDays, expected around %v", cutoff, expectedMinPast)
	}
}

func TestAlarmProtectionSafety(t *testing.T) {
	db, _, svc, _, _, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	oldTime := time.Now().AddDate(0, 0, -120)
	devID := uint(1)

	// 1. Create ACTIVE alarm (MUST NEVER BE DELETED)
	db.Create(&model.Alarm{
		DeviceID:    &devID,
		Status:      model.AlarmActive,
		TriggeredAt: oldTime,
		AlarmCode:   "ALM_ACTIVE",
		Message:     "Critical Pressure High",
	})

	// 2. Create ACKNOWLEDGED alarm (MUST NEVER BE DELETED)
	db.Create(&model.Alarm{
		DeviceID:    &devID,
		Status:      model.AlarmAcked,
		TriggeredAt: oldTime,
		AlarmCode:   "ALM_ACKED",
		Message:     "High Temperature Warning",
	})

	// 3. Create CLEARED alarm (Candidate for deletion)
	db.Create(&model.Alarm{
		DeviceID:    &devID,
		Status:      model.AlarmCleared,
		TriggeredAt: oldTime,
		ClearedAt:   &oldTime,
		AlarmCode:   "ALM_CLEARED",
		Message:     "Resolved Sensor Disconnect",
	})

	// Enable policy for execution and verify RequireBackup is false
	policy, _ := svc.GetPolicyByID("pol_cleared_alarms")
	policy.Enabled = true
	policy.RequireBackup = false
	_ = svc.SavePolicy(policy, "admin")

	execLog, err := svc.ExecutePolicy(context.Background(), "pol_cleared_alarms", "test", false)
	if err != nil {
		t.Fatalf("cleared alarms policy execution failed: %v", err)
	}

	if execLog.DeletedCount != 1 {
		t.Errorf("expected 1 cleared alarm deleted, got %d", execLog.DeletedCount)
	}

	// Verify active and acknowledged alarms remain intact
	var remainingAlarms []model.Alarm
	db.Find(&remainingAlarms)
	if len(remainingAlarms) != 2 {
		t.Fatalf("expected 2 protected alarms to remain, got %d", len(remainingAlarms))
	}

	for _, a := range remainingAlarms {
		if a.Status == model.AlarmCleared {
			t.Errorf("cleared alarm was not deleted")
		}
		if a.Status != model.AlarmActive && a.Status != model.AlarmAcked {
			t.Errorf("unexpected alarm status remained: %s", a.Status)
		}
	}
}

func TestAggregationSeparationSafety(t *testing.T) {
	db, _, svc, _, _, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	oldTime := time.Now().AddDate(0, 0, -90)

	// Seed INTERNAL_RAW aggregation
	db.Create(&model.AggregationResult{
		AggregationDefinitionID: 1,
		DeviceID:                1,
		ParameterID:             1,
		SourceType:              model.SourceInternalRaw,
		PeriodStart:             oldTime,
		PeriodEnd:               oldTime.Add(time.Hour),
		Identifier:              "20260101000000",
		Function:                model.FunctionAvg,
		IntervalSeconds:         3600,
	})

	// Seed CUSTOMER_PROCESSED aggregation (MUST NOT be deleted by internal policy!)
	db.Create(&model.AggregationResult{
		AggregationDefinitionID: 2,
		DeviceID:                1,
		ParameterID:             1,
		SourceType:              model.SourceCustomerProcessed,
		PeriodStart:             oldTime,
		PeriodEnd:               oldTime.Add(time.Hour),
		Identifier:              "20260101000000",
		Function:                model.FunctionAvg,
		IntervalSeconds:         3600,
	})

	// Enable internal aggregations policy and disable require_backup for isolated test
	policy, _ := svc.GetPolicyByID("pol_internal_agg")
	policy.Enabled = true
	policy.RequireBackup = false
	policy.RetentionDays = 60
	_ = svc.SavePolicy(policy, "test")

	execLog, err := svc.ExecutePolicy(context.Background(), "pol_internal_agg", "test", false)
	if err != nil {
		t.Fatalf("internal agg policy execution failed: %v", err)
	}
	if execLog.DeletedCount != 1 {
		t.Errorf("expected 1 internal aggregate deleted, got %d", execLog.DeletedCount)
	}

	// Verify customer aggregation is untouched
	var remainingCustomerAggs []model.AggregationResult
	db.Where("source_type = ?", model.SourceCustomerProcessed).Find(&remainingCustomerAggs)
	if len(remainingCustomerAggs) != 1 {
		t.Errorf("customer processed aggregate was improperly deleted! Expected 1, got %d", len(remainingCustomerAggs))
	}
}

func TestBackupCoverageSafetyGate(t *testing.T) {
	db, _, svc, _, backupRepo, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	oldTime := time.Now().AddDate(0, 0, -45)
	for i := 1; i <= 5; i++ {
		db.Create(&model.RawData{
			RecordUUID: fmt.Sprintf("bkp-raw-%d", i),
			DeviceID:   1,
			ReceivedAt: oldTime,
		})
	}

	// Enable policy with RequireBackup = true
	policy, _ := svc.GetPolicyByID("pol_raw_telemetry")
	policy.Enabled = true
	policy.RequireBackup = true
	policy.RequireRollup = false
	policy.RetentionDays = 30
	_ = svc.SavePolicy(policy, "test")

	// 1. Without any backup in catalog, execution MUST BE BLOCKED
	log, err := svc.ExecutePolicy(context.Background(), "pol_raw_telemetry", "test", false)
	if err == nil {
		t.Fatalf("expected execution to be blocked when no backup exists, got nil")
	}
	if log.Status != model.RetentionStatusBlocked {
		t.Errorf("expected status BLOCKED, got %s", log.Status)
	}

	// Verify no data was deleted
	var count int64
	db.Model(&model.RawData{}).Count(&count)
	if count != 5 {
		t.Errorf("data was deleted despite missing backup! Remaining: %d", count)
	}

	// 2. Create a verified backup snapshot covering the period
	tmpArchive, _ := os.CreateTemp("", "mock_backup_*.tar.gz")
	_, _ = tmpArchive.WriteString("mock backup archive content")
	_ = tmpArchive.Close()
	defer os.Remove(tmpArchive.Name())

	backupRecord := &model.BackupRecord{
		ID:               "bkp_verified_test",
		Filename:         filepath.Base(tmpArchive.Name()),
		FilePath:         tmpArchive.Name(),
		SizeBytes:        1024,
		ValidationStatus: model.ValidationStatusValid,
		Status:           model.BackupStatusCompleted,
		CreatedAt:        time.Now().UTC(),
	}
	_ = backupRepo.Create(backupRecord)

	// Now execution should succeed
	log2, err2 := svc.ExecutePolicy(context.Background(), "pol_raw_telemetry", "test", false)
	if err2 != nil {
		t.Fatalf("expected execution to succeed after backup verification, got: %v", err2)
	}
	if log2.Status != model.RetentionStatusCompleted {
		t.Errorf("expected COMPLETED status, got %s", log2.Status)
	}
	if log2.DeletedCount != 5 {
		t.Errorf("expected 5 deleted records, got %d", log2.DeletedCount)
	}
}

func TestRollupCoverageSafetyGate(t *testing.T) {
	db, _, svc, _, _, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	policy, _ := svc.GetPolicyByID("pol_raw_telemetry")
	policy.Enabled = true
	policy.RequireBackup = false
	policy.RequireRollup = true
	policy.RetentionDays = 30
	policy.ProtectedPeriodDays = 0 // Remove extra buffer so cutoff = now - 30 days
	_ = svc.SavePolicy(policy, "test")

	cutoff := svc.CalculateCutoff(policy)
	oldTime := cutoff.Add(-5 * 24 * time.Hour)

	// Seed raw data older than cutoff
	for i := 1; i <= 5; i++ {
		db.Create(&model.RawData{
			RecordUUID: fmt.Sprintf("rollup-raw-%d", i),
			DeviceID:   1,
			ReceivedAt: oldTime,
		})
	}

	// Create active aggregation definition whose last calculation is BEHIND cutoff
	staleCalcTime := cutoff.Add(-48 * time.Hour)
	db.Create(&model.AggregationDefinition{
		Name:             "Hourly Rollup",
		Code:             "DEV1_TEMP_HOURLY",
		DeviceID:         1,
		ParameterID:      1,
		Function:         model.FunctionAvg,
		Enabled:          true,
		LastCalculatedAt: &staleCalcTime,
	})

	// 1. Should be blocked because rollup has not caught up to cutoff
	log, err := svc.ExecutePolicy(context.Background(), "pol_raw_telemetry", "test", false)
	if err == nil {
		t.Fatalf("expected execution to be blocked when rollup is lagging, got nil")
	}
	if log.Status != model.RetentionStatusBlocked {
		t.Errorf("expected status BLOCKED, got %s", log.Status)
	}

	// 2. Update definition to have calculated up to or beyond cutoff
	caughtUpTime := cutoff.Add(1 * time.Hour)
	db.Model(&model.AggregationDefinition{}).Where("enabled = ?", true).Update("last_calculated_at", caughtUpTime)

	// Now execution should succeed
	log2, err2 := svc.ExecutePolicy(context.Background(), "pol_raw_telemetry", "test", false)
	if err2 != nil {
		t.Fatalf("expected execution to succeed after rollup caught up, got: %v", err2)
	}
	if log2.Status != model.RetentionStatusCompleted {
		t.Errorf("expected status COMPLETED, got %s", log2.Status)
	}
	if log2.DeletedCount != 5 {
		t.Errorf("expected 5 deleted records, got %d", log2.DeletedCount)
	}
}

func TestStorageOverviewAndCapacity(t *testing.T) {
	db, _, svc, _, _, _, cleanup := setupTestRetentionEnvironment(t)
	defer cleanup()

	devID := uint(1)
	// Seed some records in various tables
	db.Create(&model.RawData{RecordUUID: "stor-1", DeviceID: 1, ReceivedAt: time.Now()})
	db.Create(&model.SystemLog{Level: "INFO", Message: "Storage test log"})
	db.Create(&model.Alarm{DeviceID: &devID, Status: model.AlarmActive, AlarmCode: "TEST", Message: "Test Alarm"})

	overview, err := svc.GetStorageOverview()
	if err != nil {
		t.Fatalf("failed getting storage overview: %v", err)
	}

	if overview == nil {
		t.Fatalf("expected non-nil storage overview")
	}

	if len(overview.Tables) == 0 {
		t.Errorf("expected tables list in storage overview")
	}

	foundRaw := false
	for _, tInfo := range overview.Tables {
		if tInfo.TableName == "raw_data" {
			foundRaw = true
			if tInfo.RowCount < 1 {
				t.Errorf("expected at least 1 row in raw_data, got %d", tInfo.RowCount)
			}
		}
	}
	if !foundRaw {
		t.Errorf("raw_data table not found in storage overview")
	}

	if overview.CapacityStatus != "HEALTHY" && overview.CapacityStatus != "WARNING" && overview.CapacityStatus != "CRITICAL" {
		t.Errorf("invalid capacity status: %s", overview.CapacityStatus)
	}
}
