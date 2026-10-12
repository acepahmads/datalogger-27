package service

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/model"
	"datalogger/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestAlarmEnvironment(t *testing.T) (*gorm.DB, *AlarmService, *repository.AlarmRepository, *repository.SystemRepository) {
	dsn := fmt.Sprintf("file:mem_alarm_db_%d?mode=memory&cache=shared", time.Now().UnixNano()+int64(rand.Intn(10000)))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test sqlite db: %v", err)
	}

	err = db.AutoMigrate(
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.Alarm{},
		&model.AuditTrail{},
		&model.SystemLog{},
		&model.CommunicationLog{},
		&model.RetentionPolicy{},
		&model.RetentionExecutionLog{},
	)
	if err != nil {
		t.Fatalf("failed migrating test models: %v", err)
	}

	alarmRepo := repository.NewAlarmRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	alarmService := NewAlarmService(alarmRepo, systemRepo)

	return db, alarmService, alarmRepo, systemRepo
}

func TestAlarmService_CreateAndRetrieve(t *testing.T) {
	db, svc, _, _ := setupTestAlarmEnvironment(t)
	ctx := context.Background()

	// Seed Device & Parameter
	dev := model.Device{DeviceCode: "DEV-01", DeviceName: "Test Device", Status: model.DeviceStatusActive}
	db.Create(&dev)

	param := model.Parameter{DeviceID: dev.ID, ParameterCode: "temp", ParameterName: "Temperature"}
	db.Create(&param)

	trigVal := 88.5
	threshVal := 80.0
	alarm := &model.Alarm{
		DeviceID:       &dev.ID,
		ParameterID:    &param.ID,
		AlarmCode:      "HIGH_TEMP",
		Severity:       model.SeverityCritical,
		Message:        "Temperature exceeded critical limit",
		TriggerValue:   &trigVal,
		ThresholdValue: &threshVal,
		Status:         model.AlarmActive,
	}

	err := svc.CreateAlarm(ctx, alarm)
	if err != nil {
		t.Fatalf("unexpected error creating alarm: %v", err)
	}
	if alarm.ID == 0 {
		t.Fatalf("expected non-zero alarm ID, got 0")
	}
	if alarm.TriggeredAt.IsZero() {
		t.Fatalf("expected TriggeredAt to be populated automatically")
	}

	// Retrieve by ID
	retrieved, err := svc.GetAlarmByID(ctx, alarm.ID)
	if err != nil {
		t.Fatalf("failed retrieving alarm: %v", err)
	}
	if retrieved.AlarmCode != "HIGH_TEMP" {
		t.Errorf("expected AlarmCode HIGH_TEMP, got %s", retrieved.AlarmCode)
	}
	if retrieved.Severity != model.SeverityCritical {
		t.Errorf("expected Severity CRITICAL, got %s", retrieved.Severity)
	}
	if retrieved.Status != model.AlarmActive {
		t.Errorf("expected Status ACTIVE, got %s", retrieved.Status)
	}
	if retrieved.Device == nil || retrieved.Device.DeviceCode != "DEV-01" {
		t.Errorf("expected preloaded Device with code DEV-01")
	}
	if retrieved.Parameter == nil || retrieved.Parameter.ParameterCode != "temp" {
		t.Errorf("expected preloaded Parameter with code temp")
	}
}

func TestAlarmService_ListFilteringAndPagination(t *testing.T) {
	db, svc, _, _ := setupTestAlarmEnvironment(t)
	ctx := context.Background()

	dev := model.Device{DeviceCode: "DEV-FILTER", DeviceName: "Filter Test Device"}
	db.Create(&dev)

	// Seed 5 alarms with different severities and statuses
	alarms := []*model.Alarm{
		{DeviceID: &dev.ID, AlarmCode: "ALM-1", Severity: model.SeverityInfo, Status: model.AlarmActive, Message: "Info event 1"},
		{DeviceID: &dev.ID, AlarmCode: "ALM-2", Severity: model.SeverityWarning, Status: model.AlarmActive, Message: "Warning event 2"},
		{DeviceID: &dev.ID, AlarmCode: "ALM-3", Severity: model.SeverityCritical, Status: model.AlarmAcked, Message: "Critical event 3"},
		{DeviceID: &dev.ID, AlarmCode: "ALM-4", Severity: model.SeverityWarning, Status: model.AlarmCleared, Message: "Warning cleared 4"},
		{DeviceID: &dev.ID, AlarmCode: "ALM-5", Severity: model.SeverityCritical, Status: model.AlarmCleared, Message: "Critical cleared 5"},
	}
	for _, a := range alarms {
		if err := svc.CreateAlarm(ctx, a); err != nil {
			t.Fatalf("failed creating test alarm: %v", err)
		}
	}

	// Filter by Status = ACTIVE
	activeList, totalActive, err := svc.ListAlarms(ctx, repository.AlarmFilterParams{Status: "ACTIVE"})
	if err != nil {
		t.Fatalf("failed listing active alarms: %v", err)
	}
	if totalActive != 2 || len(activeList) != 2 {
		t.Errorf("expected 2 active alarms, got %d (total: %d)", len(activeList), totalActive)
	}

	// Filter by Severity = CRITICAL
	critList, totalCrit, err := svc.ListAlarms(ctx, repository.AlarmFilterParams{Severity: "CRITICAL"})
	if err != nil {
		t.Fatalf("failed listing critical alarms: %v", err)
	}
	if totalCrit != 2 || len(critList) != 2 {
		t.Errorf("expected 2 critical alarms, got %d (total: %d)", len(critList), totalCrit)
	}

	// Search filter
	searchList, totalSearch, err := svc.ListAlarms(ctx, repository.AlarmFilterParams{Search: "cleared"})
	if err != nil {
		t.Fatalf("failed searching alarms: %v", err)
	}
	if totalSearch != 2 || len(searchList) != 2 {
		t.Errorf("expected 2 search results for 'cleared', got %d", len(searchList))
	}

	// Pagination: Page 1, PageSize 2
	p1List, totalP1, err := svc.ListAlarms(ctx, repository.AlarmFilterParams{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("failed paginated query: %v", err)
	}
	if totalP1 != 5 || len(p1List) != 2 {
		t.Errorf("expected page 1 with 2 items from total 5, got %d (total: %d)", len(p1List), totalP1)
	}
}

func TestAlarmService_LifecycleTransitions(t *testing.T) {
	_, svc, _, _ := setupTestAlarmEnvironment(t)
	ctx := context.Background()

	alarm := &model.Alarm{
		AlarmCode: "LIFECYCLE-01",
		Severity:  model.SeverityWarning,
		Message:   "Pressure dropping",
		Status:    model.AlarmActive,
	}
	if err := svc.CreateAlarm(ctx, alarm); err != nil {
		t.Fatalf("failed creating alarm: %v", err)
	}

	// 1. Transition: ACTIVE -> ACKNOWLEDGED
	acked, err := svc.AcknowledgeAlarm(ctx, alarm.ID, "operator_john")
	if err != nil {
		t.Fatalf("failed acknowledging alarm: %v", err)
	}
	if acked.Status != model.AlarmAcked {
		t.Errorf("expected status ACKNOWLEDGED, got %s", acked.Status)
	}
	if acked.AcknowledgedBy != "operator_john" {
		t.Errorf("expected acknowledged_by operator_john, got %s", acked.AcknowledgedBy)
	}
	if acked.AcknowledgedAt == nil || acked.AcknowledgedAt.IsZero() {
		t.Errorf("expected acknowledged_at timestamp to be set")
	}

	// 2. Idempotent acknowledge on already ACKNOWLEDGED alarm
	reAcked, err := svc.AcknowledgeAlarm(ctx, alarm.ID, "operator_john")
	if err != nil {
		t.Fatalf("expected idempotent success on re-acknowledgement, got error: %v", err)
	}
	if reAcked.Status != model.AlarmAcked {
		t.Errorf("expected status still ACKNOWLEDGED, got %s", reAcked.Status)
	}

	// 3. Transition: ACKNOWLEDGED -> CLEARED
	cleared, err := svc.ClearAlarm(ctx, alarm.ID)
	if err != nil {
		t.Fatalf("failed clearing alarm: %v", err)
	}
	if cleared.Status != model.AlarmCleared {
		t.Errorf("expected status CLEARED, got %s", cleared.Status)
	}
	if cleared.ClearedAt == nil || cleared.ClearedAt.IsZero() {
		t.Errorf("expected cleared_at timestamp to be set")
	}
}

func TestAlarmService_InvalidLifecycleTransitions(t *testing.T) {
	_, svc, _, _ := setupTestAlarmEnvironment(t)
	ctx := context.Background()

	// 1. Validation on creation
	badSeverity := &model.Alarm{AlarmCode: "BAD-1", Severity: "INVALID_SEV", Message: "Test"}
	if err := svc.CreateAlarm(ctx, badSeverity); err != repository.ErrInvalidAlarmSeverity {
		t.Errorf("expected ErrInvalidAlarmSeverity, got %v", err)
	}

	emptyCode := &model.Alarm{AlarmCode: "", Severity: model.SeverityWarning, Message: "Test"}
	if err := svc.CreateAlarm(ctx, emptyCode); err != repository.ErrMissingAlarmCode {
		t.Errorf("expected ErrMissingAlarmCode, got %v", err)
	}

	emptyMsg := &model.Alarm{AlarmCode: "NO-MSG", Severity: model.SeverityWarning, Message: ""}
	if err := svc.CreateAlarm(ctx, emptyMsg); err != repository.ErrMissingAlarmMessage {
		t.Errorf("expected ErrMissingAlarmMessage, got %v", err)
	}

	// 2. Create alarm and clear it
	alarm := &model.Alarm{AlarmCode: "ALM-INVALID", Severity: model.SeverityCritical, Message: "Critical pump failure"}
	if err := svc.CreateAlarm(ctx, alarm); err != nil {
		t.Fatalf("failed creating alarm: %v", err)
	}

	if _, err := svc.ClearAlarm(ctx, alarm.ID); err != nil {
		t.Fatalf("failed clearing alarm: %v", err)
	}

	// 3. Invalid transition: Cannot acknowledge an already CLEARED alarm
	_, err := svc.AcknowledgeAlarm(ctx, alarm.ID, "admin")
	if err != repository.ErrCannotAcknowledgeCleared {
		t.Errorf("expected ErrCannotAcknowledgeCleared, got %v", err)
	}

	// 4. Invalid transition: Cannot clear an already CLEARED alarm
	_, err = svc.ClearAlarm(ctx, alarm.ID)
	if err != repository.ErrAlarmAlreadyCleared {
		t.Errorf("expected ErrAlarmAlreadyCleared, got %v", err)
	}

	// 5. Non-existent alarm
	_, err = svc.GetAlarmByID(ctx, 999999)
	if err != repository.ErrAlarmNotFound {
		t.Errorf("expected ErrAlarmNotFound, got %v", err)
	}
}

func TestAlarmService_AuditLogging(t *testing.T) {
	_, svc, _, systemRepo := setupTestAlarmEnvironment(t)
	ctx := context.Background()

	alarm := &model.Alarm{
		AlarmCode: "AUDIT-01",
		Severity:  model.SeverityWarning,
		Message:   "Flow meter low warning",
	}
	if err := svc.CreateAlarm(ctx, alarm); err != nil {
		t.Fatalf("failed creating alarm: %v", err)
	}

	_, _ = svc.AcknowledgeAlarm(ctx, alarm.ID, "supervisor_bob")
	_, _ = svc.ClearAlarm(ctx, alarm.ID)

	trails, err := systemRepo.GetAuditTrails(10)
	if err != nil {
		t.Fatalf("failed getting audit trails: %v", err)
	}

	// Verify that ALARM_CREATE, ALARM_ACKNOWLEDGE, and ALARM_CLEAR are logged
	actionsFound := make(map[string]bool)
	for _, trail := range trails {
		if trail.Resource == "ALARM" {
			actionsFound[trail.Action] = true
		}
	}

	if !actionsFound["ALARM_CREATE"] {
		t.Errorf("expected ALARM_CREATE audit trail entry")
	}
	if !actionsFound["ALARM_ACKNOWLEDGE"] {
		t.Errorf("expected ALARM_ACKNOWLEDGE audit trail entry")
	}
	if !actionsFound["ALARM_CLEAR"] {
		t.Errorf("expected ALARM_CLEAR audit trail entry")
	}
}

func TestAlarmService_RetentionProtectionIntegration(t *testing.T) {
	db, svc, _, systemRepo := setupTestAlarmEnvironment(t)
	ctx := context.Background()

	cfg := &config.Config{
		DataDir:   t.TempDir(),
		BackupDir: t.TempDir(),
	}

	retentionRepo := repository.NewRetentionRepository(db)
	backupRepo := repository.NewBackupRepository(db)
	retentionService := NewRetentionService(db, cfg, retentionRepo, backupRepo, systemRepo, nil, nil)

	// Seed 3 alarms:
	// Alarm 1: ACTIVE (old date) - MUST NOT BE PRUNED
	// Alarm 2: ACKNOWLEDGED (old date) - MUST NOT BE PRUNED
	// Alarm 3: CLEARED (old date) - MAY BE PRUNED
	oldTime := time.Now().Add(-120 * 24 * time.Hour)

	alm1 := &model.Alarm{
		AlarmCode:   "PROT-ACTIVE",
		Severity:    model.SeverityCritical,
		Message:     "Active old alarm",
		Status:      model.AlarmActive,
		TriggeredAt: oldTime,
	}
	_ = svc.CreateAlarm(ctx, alm1)
	db.Model(alm1).Update("created_at", oldTime)

	alm2 := &model.Alarm{
		AlarmCode:   "PROT-ACKED",
		Severity:    model.SeverityWarning,
		Message:     "Acknowledged old alarm",
		Status:      model.AlarmAcked,
		TriggeredAt: oldTime,
	}
	_ = svc.CreateAlarm(ctx, alm2)
	db.Model(alm2).Update("created_at", oldTime)

	alm3 := &model.Alarm{
		AlarmCode:   "PROT-CLEARED",
		Severity:    model.SeverityInfo,
		Message:     "Cleared old alarm",
		Status:      model.AlarmCleared,
		TriggeredAt: oldTime,
		ClearedAt:   &oldTime,
	}
	_ = svc.CreateAlarm(ctx, alm3)
	db.Model(alm3).Update("created_at", oldTime)

	// Enable policy for execution and verify RequireBackup is false
	policy, err := retentionService.GetPolicyByID("pol_cleared_alarms")
	if err != nil {
		t.Fatalf("failed retrieving pol_cleared_alarms: %v", err)
	}
	policy.Enabled = true
	policy.RequireBackup = false
	_ = retentionService.SavePolicy(policy, "admin")

	// Execute cleared alarms retention policy
	_, err = retentionService.ExecutePolicy(ctx, "pol_cleared_alarms", "test", false)
	if err != nil {
		t.Fatalf("retention execution failed: %v", err)
	}

	// Verify Alarm 1 and Alarm 2 still exist
	var check1, check2, check3 model.Alarm
	err1 := db.First(&check1, alm1.ID).Error
	err2 := db.First(&check2, alm2.ID).Error
	err3 := db.First(&check3, alm3.ID).Error

	if err1 != nil {
		t.Errorf("ACTIVE alarm was deleted by retention: %v", err1)
	}
	if err2 != nil {
		t.Errorf("ACKNOWLEDGED alarm was deleted by retention: %v", err2)
	}
	if err3 == nil {
		t.Errorf("CLEARED alarm older than 30 days was not pruned by retention")
	}
}
