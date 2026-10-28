package service_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestEvaluatorEnvironment(t *testing.T) (*gorm.DB, *service.AlarmEvaluator, *service.AlarmService, *repository.AlarmRepository) {
	dsn := fmt.Sprintf("file:mem_eval_db_%d?mode=memory&cache=shared", time.Now().UnixNano()+int64(rand.Intn(100000)))
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
	)
	if err != nil {
		t.Fatalf("failed migrating test models: %v", err)
	}

	alarmRepo := repository.NewAlarmRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)

	cfg := service.DefaultAlarmEvaluatorConfig()
	cfg.DefaultHysteresisPercent = 1.0 // 1% deadband

	evaluator := service.NewAlarmEvaluator(context.Background(), alarmService, cfg, nil)

	return db, evaluator, alarmService, alarmRepo
}

func createTestDeviceAndParam(db *gorm.DB, wLow, wHigh, cLow, cHigh *float64, hyst *float64) (*model.Device, *model.Parameter) {
	dev := model.Device{
		DeviceCode: fmt.Sprintf("DEV-%d", rand.Intn(100000)),
		DeviceName: "Test Compressor",
		Status:     model.DeviceStatusActive,
	}
	db.Create(&dev)

	param := model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "pressure",
		ParameterName: "Discharge Pressure",
		Unit:          "bar",
		DataType:      model.DataTypeFloat32,
		Enabled:       true,
		WarningLow:    wLow,
		WarningHigh:   wHigh,
		CriticalLow:   cLow,
		CriticalHigh:  cHigh,
		Hysteresis:    hyst,
	}
	db.Create(&param)

	return &dev, &param
}

// 1. Test normal values within safe boundaries
func TestAlarmEvaluator_NormalValues(t *testing.T) {
	db, evaluator, _, _ := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wL, wH, cL, cH := 10.0, 80.0, 5.0, 90.0
	dev, param := createTestDeviceAndParam(db, &wL, &wH, &cL, &cH, nil)

	telemetry := &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		Unit:           param.Unit,
		ProcessedValue: 50.0, // Normal midpoint
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}

	evaluator.EvaluateTelemetry(ctx, telemetry, param)

	if evaluator.GetActiveCount() != 0 {
		t.Fatalf("expected 0 active alarms for normal value, got %d", evaluator.GetActiveCount())
	}
}

// 2. Test Warning High & Warning Low triggers
func TestAlarmEvaluator_WarningThresholds(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wL, wH, cL, cH := 20.0, 80.0, 10.0, 90.0
	dev, param := createTestDeviceAndParam(db, &wL, &wH, &cL, &cH, nil)

	// Trigger Warning High (82.5 >= 80.0)
	tHigh := &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		Unit:           param.Unit,
		ProcessedValue: 82.5,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}
	evaluator.EvaluateTelemetry(ctx, tHigh, param)

	active, err := alarmRepo.GetActive(ctx)
	if err != nil || len(active) != 1 {
		t.Fatalf("expected 1 active warning high alarm, got %d (err: %v)", len(active), err)
	}
	if active[0].AlarmCode != service.AlarmCodeWarningHigh || active[0].Severity != model.SeverityWarning {
		t.Errorf("expected WARNING_HIGH with WARNING severity, got %s / %s", active[0].AlarmCode, active[0].Severity)
	}

	// Trigger Warning Low on a separate parameter (18.0 <= 20.0)
	_, paramLow := createTestDeviceAndParam(db, &wL, &wH, &cL, &cH, nil)
	tLow := &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    paramLow.ID,
		ParameterCode:  paramLow.ParameterCode,
		ParameterName:  paramLow.ParameterName,
		Unit:           paramLow.Unit,
		ProcessedValue: 18.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}
	evaluator.EvaluateTelemetry(ctx, tLow, paramLow)

	active, _ = alarmRepo.GetActive(ctx)
	if len(active) != 2 {
		t.Fatalf("expected 2 active alarms, got %d", len(active))
	}
}

// 3. Test Critical High & Critical Low with Precedence
func TestAlarmEvaluator_CriticalPrecedence(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wL, wH, cL, cH := 20.0, 80.0, 10.0, 90.0
	dev, param := createTestDeviceAndParam(db, &wL, &wH, &cL, &cH, nil)

	// Ingest 95.0 (satisfies both Warning High 80 and Critical High 90)
	// Critical must take precedence!
	tCrit := &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		Unit:           param.Unit,
		ProcessedValue: 95.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}
	evaluator.EvaluateTelemetry(ctx, tCrit, param)

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Fatalf("expected exactly 1 active critical alarm, got %d", len(active))
	}
	if active[0].AlarmCode != service.AlarmCodeCriticalHigh || active[0].Severity != model.SeverityCritical {
		t.Errorf("expected CRITICAL_HIGH, got %s / %s", active[0].AlarmCode, active[0].Severity)
	}
}

// 4. Test Transition from Warning to Critical (Escalation)
func TestAlarmEvaluator_EscalationWarningToCritical(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wL, wH, cL, cH := 20.0, 80.0, 10.0, 90.0
	dev, param := createTestDeviceAndParam(db, &wL, &wH, &cL, &cH, nil)

	// Step 1: Warning High (85.0)
	tWarn := &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		Unit:           param.Unit,
		ProcessedValue: 85.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}
	evaluator.EvaluateTelemetry(ctx, tWarn, param)

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeWarningHigh {
		t.Fatalf("expected 1 WARNING_HIGH alarm initially, got %d", len(active))
	}
	firstAlarmID := active[0].ID

	// Step 2: Value escalates to Critical High (92.0)
	tCrit := &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		Unit:           param.Unit,
		ProcessedValue: 92.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}
	evaluator.EvaluateTelemetry(ctx, tCrit, param)

	active, _ = alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Fatalf("expected exactly 1 active alarm after escalation, got %d", len(active))
	}
	if active[0].AlarmCode != service.AlarmCodeCriticalHigh {
		t.Errorf("expected escalated CRITICAL_HIGH alarm, got %s", active[0].AlarmCode)
	}

	// Verify old warning alarm was CLEARED
	var oldAlarm model.Alarm
	db.First(&oldAlarm, firstAlarmID)
	if oldAlarm.Status != model.AlarmCleared {
		t.Errorf("expected previous warning alarm %d to be CLEARED, got %s", firstAlarmID, oldAlarm.Status)
	}
}

// 5. Test De-escalation from Critical to Warning (Recovery without premature clearing)
func TestAlarmEvaluator_DeescalationCriticalToWarning(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wL, wH, cL, cH := 20.0, 80.0, 10.0, 90.0
	hyst := 1.0 // 1.0 bar explicit hysteresis
	dev, param := createTestDeviceAndParam(db, &wL, &wH, &cL, &cH, &hyst)

	// Step 1: Ingest Critical High (95.0)
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 95.0, Quality: model.QualityGood,
	}, param)

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeCriticalHigh {
		t.Fatalf("expected CRITICAL_HIGH active, got %v", active)
	}

	// Step 2: Value drops to 85.0 (below CriticalHigh - Hysteresis = 90 - 1 = 89, but still above WarningHigh = 80)
	// Must de-escalate to WARNING_HIGH, NOT clear completely to normal!
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 85.0, Quality: model.QualityGood,
	}, param)

	active, _ = alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Fatalf("expected 1 active alarm after de-escalation, got %d", len(active))
	}
	if active[0].AlarmCode != service.AlarmCodeWarningHigh {
		t.Errorf("expected de-escalated WARNING_HIGH alarm, got %s", active[0].AlarmCode)
	}
}

// 6. Test Hysteresis Deadband Recovery
func TestAlarmEvaluator_HysteresisDeadband(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wH := 80.0
	hyst := 2.0 // Hysteresis is 2.0: Trigger >= 80.0, Recover < 78.0
	dev, param := createTestDeviceAndParam(db, nil, &wH, nil, nil, &hyst)

	// Step 1: Trigger warning at 81.0
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 81.0, Quality: model.QualityGood,
	}, param)

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Fatalf("expected 1 active alarm at 81.0")
	}

	// Step 2: Value drops to 79.0 (Below 80.0 threshold, but above 78.0 hysteresis floor)
	// Must STAY ACTIVE (latched) to prevent fluttering!
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 79.0, Quality: model.QualityGood,
	}, param)

	active, _ = alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Fatalf("expected alarm to remain LATCHED in hysteresis deadband at 79.0, but got %d active", len(active))
	}

	// Step 3: Value drops to 77.5 (Below 78.0 hysteresis floor)
	// Now it MUST CLEAR!
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 77.5, Quality: model.QualityGood,
	}, param)

	active, _ = alarmRepo.GetActive(ctx)
	if len(active) != 0 {
		t.Fatalf("expected alarm to be CLEARED past hysteresis floor at 77.5, but got %d active", len(active))
	}
}

// 7. Test Repeated Samples without Duplicate Alarms
func TestAlarmEvaluator_RepeatedSamplesAntiDuplication(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wH := 80.0
	dev, param := createTestDeviceAndParam(db, nil, &wH, nil, nil, nil)

	// 10 consecutive readings in alarm state
	for i := 0; i < 10; i++ {
		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 85.0 + float64(i)*0.1, Quality: model.QualityGood,
		}, param)
	}

	// In memory and DB, there must be EXACTLY ONE active alarm!
	if evaluator.GetActiveCount() != 1 {
		t.Errorf("expected 1 in-memory active alarm, got %d", evaluator.GetActiveCount())
	}

	var totalAlarms int64
	db.Model(&model.Alarm{}).Count(&totalAlarms)
	if totalAlarms != 1 {
		t.Errorf("expected exactly 1 alarm record in database, got %d", totalAlarms)
	}

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Errorf("expected exactly 1 active alarm in database, got %d", len(active))
	}
}

// 8. Test Concurrent Telemetry Evaluations (Thread-Safety)
func TestAlarmEvaluator_ConcurrentTelemetrySafety(t *testing.T) {
	db, evaluator, _, _ := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wH := 80.0
	dev, param := createTestDeviceAndParam(db, nil, &wH, nil, nil, nil)

	var wg sync.WaitGroup
	workers := 20
	iterations := 20

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
					DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 85.5, Quality: model.QualityGood,
				}, param)
			}
		}(w)
	}

	wg.Wait()

	// Despite 400 concurrent evaluations, only 1 alarm must exist
	var totalAlarms int64
	db.Model(&model.Alarm{}).Count(&totalAlarms)
	if totalAlarms != 1 {
		t.Errorf("concurrency violation: expected exactly 1 alarm record, got %d", totalAlarms)
	}
	if evaluator.GetActiveCount() != 1 {
		t.Errorf("concurrency violation: expected 1 in-memory active entry, got %d", evaluator.GetActiveCount())
	}
}

// 9. Test Data Quality: NaN, Inf, and Bad Quality rejection
func TestAlarmEvaluator_DataQualityRejection(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wH := 80.0
	dev, param := createTestDeviceAndParam(db, nil, &wH, nil, nil, nil)

	// NaN value: rejected
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: math.NaN(), Quality: model.QualityGood,
	}, param)

	// Positive Inf: rejected
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: math.Inf(1), Quality: model.QualityGood,
	}, param)

	// QualityBad: rejected
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 99.9, Quality: model.QualityBad,
	}, param)

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 0 {
		t.Errorf("expected 0 alarms from invalid or bad quality telemetry, got %d", len(active))
	}
}

// 10. Test Uncertain Quality Annotation
func TestAlarmEvaluator_UncertainQualityAnnotation(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wH := 80.0
	dev, param := createTestDeviceAndParam(db, nil, &wH, nil, nil, nil)

	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		ParameterID:    param.ID,
		ProcessedValue: 85.0,
		Quality:        model.QualityUncertain,
		QualityReason:  model.ReasonHoldAnomalyGrace,
	}, param)

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Fatalf("expected 1 alarm triggered for UNCERTAIN quality, got %d", len(active))
	}
	if !contains(active[0].Message, "UNCERTAIN") {
		t.Errorf("expected message to note UNCERTAIN quality, got: %s", active[0].Message)
	}
}

// 11. Test Startup State Synchronization from Database (Restart / Reload behavior)
func TestAlarmEvaluator_StartupReloadFromDB(t *testing.T) {
	db, _, alarmService, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wH := 80.0
	dev, param := createTestDeviceAndParam(db, nil, &wH, nil, nil, nil)

	// Seed 1 ACTIVE alarm directly into DB (representing pre-existing un-cleared alarm prior to restart)
	trigVal := 88.0
	seedAlarm := &model.Alarm{
		DeviceID:       &dev.ID,
		ParameterID:    &param.ID,
		AlarmCode:      service.AlarmCodeWarningHigh,
		Severity:       model.SeverityWarning,
		Message:        "Existing warning",
		TriggerValue:   &trigVal,
		ThresholdValue: &wH,
		Status:         model.AlarmActive,
		TriggeredAt:    time.Now().Add(-10 * time.Minute),
	}
	_ = alarmRepo.Create(ctx, seedAlarm)

	// Instantiate a NEW evaluator (simulating server reboot)
	newEvaluator := service.NewAlarmEvaluator(ctx, alarmService, service.DefaultAlarmEvaluatorConfig(), nil)

	if newEvaluator.GetActiveCount() != 1 {
		t.Fatalf("expected new evaluator to reload 1 active alarm from DB on startup, got %d", newEvaluator.GetActiveCount())
	}

	// Ingest value in warning zone: must NOT create a duplicate alarm!
	newEvaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 86.0, Quality: model.QualityGood,
	}, param)

	var totalAlarms int64
	db.Model(&model.Alarm{}).Count(&totalAlarms)
	if totalAlarms != 1 {
		t.Errorf("reboot anti-duplication failed: expected 1 alarm record in DB, got %d", totalAlarms)
	}
}

// 12. Test Acknowledged Alarm Preservation on Recovery
func TestAlarmEvaluator_AcknowledgedPreservationOnRecovery(t *testing.T) {
	db, evaluator, alarmService, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wH := 80.0
	hyst := 1.0
	dev, param := createTestDeviceAndParam(db, nil, &wH, nil, nil, &hyst)

	// 1. Trigger alarm at 85.0
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 85.0, Quality: model.QualityGood,
	}, param)

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Fatalf("expected 1 active alarm, got %d", len(active))
	}
	alarmID := active[0].ID

	// 2. Operator acknowledges the alarm with notes
	ackAlarm, err := alarmService.AcknowledgeAlarm(ctx, alarmID, "operator_alex", "Investigating pump cooling fan")
	if err != nil {
		t.Fatalf("failed acknowledging alarm: %v", err)
	}
	if ackAlarm.Status != model.AlarmAcked || ackAlarm.Notes != "Investigating pump cooling fan" {
		t.Fatalf("acknowledgement mismatch: %v", ackAlarm)
	}

	// 3. Condition physically recovers below hysteresis (75.0 < 80.0 - 1.0)
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID: dev.ID, ParameterID: param.ID, ProcessedValue: 75.0, Quality: model.QualityGood,
	}, param)

	// Alarm should be CLEARED, but acknowledged_at, acknowledged_by, and notes MUST BE PRESERVED!
	clearedAlarm, err := alarmRepo.GetByID(ctx, alarmID)
	if err != nil {
		t.Fatalf("failed getting alarm: %v", err)
	}
	if clearedAlarm.Status != model.AlarmCleared {
		t.Errorf("expected status CLEARED, got %s", clearedAlarm.Status)
	}
	if clearedAlarm.ClearedAt == nil {
		t.Errorf("expected ClearedAt timestamp to be set")
	}
	if clearedAlarm.AcknowledgedBy != "operator_alex" {
		t.Errorf("expected AcknowledgedBy operator_alex to be preserved, got %s", clearedAlarm.AcknowledgedBy)
	}
	if clearedAlarm.AcknowledgedAt == nil {
		t.Errorf("expected AcknowledgedAt timestamp to be preserved")
	}
	if clearedAlarm.Notes != "Investigating pump cooling fan" {
		t.Errorf("expected Notes to be preserved, got '%s'", clearedAlarm.Notes)
	}
}

// 13. Test TelemetryService.Ingest End-to-End Pipeline Integration
func TestTelemetryService_IngestAlarmIntegration(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	wH := 50.0
	dev, param := createTestDeviceAndParam(db, nil, &wH, nil, nil, nil)

	telemetryRepo := repository.NewTelemetryRepository(db)
	telemetryService := service.NewTelemetryService(telemetryRepo, nil, service.DefaultTelemetryConfig())
	defer telemetryService.Stop()

	// Attach evaluator to TelemetryService
	telemetryService.SetAlarmEvaluator(evaluator)

	// Ingest telemetry via TelemetryService.Ingest
	payload := &service.TelemetryIngestPayload{
		DeviceID:      dev.ID,
		DeviceCode:    dev.DeviceCode,
		DeviceName:    dev.DeviceName,
		ParameterID:   param.ID,
		ParameterCode: param.ParameterCode,
		ParameterName: param.ParameterName,
		Unit:          param.Unit,
		DataType:      param.DataType,
		RawValue:      65.0,
		Value:         65.0,
		Quality:       model.QualityGood,
		ReceivedAt:    time.Now().UTC(),
		Parameter:     param,
	}

	err := telemetryService.Ingest(payload)
	if err != nil {
		t.Fatalf("unexpected Ingest error: %v", err)
	}

	// Verify alarm was automatically triggered through TelemetryService pipeline
	active, err := alarmRepo.GetActive(ctx)
	if err != nil || len(active) != 1 {
		t.Fatalf("expected 1 active alarm triggered through Ingest pipeline, got %d (err: %v)", len(active), err)
	}
	if active[0].AlarmCode != service.AlarmCodeWarningHigh {
		t.Errorf("expected WARNING_HIGH alarm, got %s", active[0].AlarmCode)
	}
}

// -------------------------------------------------------------
// Phase 6 Wave 2 Automated Test Suite
// Task #60: Device Offline
// Task #61: Communication Error
// Task #62: Data Abnormality
// -------------------------------------------------------------

// 14. Test Device Offline Timeout, Alarm Generation, and Recovery
func TestAlarmEvaluator_DeviceOfflineTimeoutAndRecovery(t *testing.T) {
	db, evaluator, alarmService, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, _ := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	// A. Device with recent communication -> No alarm
	recent := time.Now().UTC().Add(-10 * time.Second)
	healthHealthy := &service.DeviceHealthInput{
		DeviceID:             dev.ID,
		State:                "CONNECTED",
		ConsecutiveSuccesses: 5,
		LastCommSuccess:      &recent,
	}
	evaluator.EvaluateDeviceOffline(ctx, dev, healthHealthy)
	if evaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Fatalf("expected 0 offline alarms for healthy device, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	// B. Device communication timed out (> 60s ago) -> Triggers CRITICAL DEVICE_OFFLINE alarm
	oldSuccess := time.Now().UTC().Add(-75 * time.Second)
	healthOffline := &service.DeviceHealthInput{
		DeviceID:            dev.ID,
		State:               "ERROR",
		ConsecutiveFailures: 6,
		LastCommSuccess:     &oldSuccess,
		LastErrorCategory:   "TIMEOUT",
		LastError:           "i/o timeout",
	}
	evaluator.EvaluateDeviceOffline(ctx, dev, healthOffline)

	if evaluator.GetActiveDeviceOfflineCount() != 1 {
		t.Fatalf("expected 1 active offline alarm, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	active, err := alarmRepo.GetActive(ctx)
	if err != nil || len(active) != 1 {
		t.Fatalf("expected 1 active alarm in repo, got %d (err: %v)", len(active), err)
	}
	alm := active[0]
	if alm.AlarmCode != service.AlarmCodeDeviceOffline {
		t.Errorf("expected alarm code %s, got %s", service.AlarmCodeDeviceOffline, alm.AlarmCode)
	}
	if alm.Severity != model.SeverityCritical {
		t.Errorf("expected severity CRITICAL, got %s", alm.Severity)
	}
	if !hasSubstr(alm.Message, dev.DeviceCode) || !hasSubstr(alm.Message, "TIMEOUT") {
		t.Errorf("expected message to contain device code and category, got: %s", alm.Message)
	}

	// C. Operator acknowledges the offline alarm
	acked, err := alarmService.AcknowledgeAlarm(ctx, alm.ID, "operator_bob", "Dispatched field technician")
	if err != nil || acked.Status != model.AlarmAcked {
		t.Fatalf("failed acknowledging alarm: %v", err)
	}

	// D. Device recovers: successful communication resolves the alarm while preserving ack history
	nowSuccess := time.Now().UTC()
	healthRecovered := &service.DeviceHealthInput{
		DeviceID:             dev.ID,
		State:                "CONNECTED",
		ConsecutiveSuccesses: 2,
		LastCommSuccess:      &nowSuccess,
	}
	evaluator.EvaluateDeviceCommSuccess(ctx, dev, healthRecovered)

	if evaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Errorf("expected active offline count 0 after recovery, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	clearedAlm, err := alarmRepo.GetByID(ctx, alm.ID)
	if err != nil {
		t.Fatalf("failed fetching cleared alarm: %v", err)
	}
	if clearedAlm.Status != model.AlarmCleared {
		t.Errorf("expected alarm status CLEARED, got %s", clearedAlm.Status)
	}
	if clearedAlm.AcknowledgedBy != "operator_bob" {
		t.Errorf("expected acknowledged_by preserved as operator_bob, got %s", clearedAlm.AcknowledgedBy)
	}
	if clearedAlm.Notes != "Dispatched field technician" {
		t.Errorf("expected notes preserved, got %s", clearedAlm.Notes)
	}
	if clearedAlm.ClearedAt == nil {
		t.Errorf("expected cleared_at timestamp to be set")
	}
}

// 15. Test Administrative Disablement and Maintenance Suppression
func TestAlarmEvaluator_DeviceOfflineDisabledSuppression(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, _ := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)
	dev.Enabled = false // Administratively disabled

	oldSuccess := time.Now().UTC().Add(-300 * time.Second)
	health := &service.DeviceHealthInput{
		DeviceID:        dev.ID,
		State:           "DISABLED",
		LastCommSuccess: &oldSuccess,
	}

	evaluator.EvaluateDeviceOffline(ctx, dev, health)

	if evaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Errorf("disabled device must not trigger offline alarms, got %d", evaluator.GetActiveDeviceOfflineCount())
	}
	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 0 {
		t.Errorf("expected 0 alarms in DB for disabled device, got %d", len(active))
	}
}

// 16. Test Repeated Timeout Checks without Duplicate Alarms
func TestAlarmEvaluator_DeviceOfflineRepeatedChecksNoDuplicate(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, _ := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	oldSuccess := time.Now().UTC().Add(-120 * time.Second)
	health := &service.DeviceHealthInput{
		DeviceID:        dev.ID,
		State:           "ERROR",
		LastCommSuccess: &oldSuccess,
	}

	// Repeatedly evaluate 5 times
	for i := 0; i < 5; i++ {
		evaluator.EvaluateDeviceOffline(ctx, dev, health)
	}

	active, err := alarmRepo.GetActive(ctx)
	if err != nil || len(active) != 1 {
		t.Fatalf("expected exactly 1 active alarm across repeated evaluations, got %d (err: %v)", len(active), err)
	}
}

// 17. Test Communication Error Threshold, Deduplication, and Cooldown
func TestAlarmEvaluator_CommErrorThresholdAndCooldown(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, _ := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	errTimeout := fmt.Errorf("read tcp 192.168.1.100:502: i/o timeout")

	// Failures 1 through 4: Below threshold of 5 -> No alarm
	for i := 1; i <= 4; i++ {
		health := &service.DeviceHealthInput{
			DeviceID:            dev.ID,
			State:               "DEGRADED",
			ConsecutiveFailures: i,
			LastErrorCategory:   "TIMEOUT",
			LastError:           errTimeout.Error(),
		}
		evaluator.EvaluateDeviceCommFailure(ctx, dev, health, errTimeout)
		if evaluator.GetActiveCommErrorsCount() != 0 {
			t.Fatalf("failure %d should not have triggered comm error alarm", i)
		}
	}

	// 5th failure: Meets consecutive limit -> Triggers WARNING COMMUNICATION_ERROR
	health5 := &service.DeviceHealthInput{
		DeviceID:            dev.ID,
		State:               "ERROR",
		ConsecutiveFailures: 5,
		LastErrorCategory:   "TIMEOUT",
		LastError:           errTimeout.Error(),
	}
	evaluator.EvaluateDeviceCommFailure(ctx, dev, health5, errTimeout)

	if evaluator.GetActiveCommErrorsCount() != 1 {
		t.Fatalf("expected 1 active comm error alarm at threshold, got %d", evaluator.GetActiveCommErrorsCount())
	}

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 {
		t.Fatalf("expected 1 active alarm in repo, got %d", len(active))
	}
	if active[0].AlarmCode != service.AlarmCodeCommError {
		t.Errorf("expected alarm code %s, got %s", service.AlarmCodeCommError, active[0].AlarmCode)
	}
	if active[0].Severity != model.SeverityWarning {
		t.Errorf("expected severity WARNING, got %s", active[0].Severity)
	}

	// 6th and 7th failures: Cooldown / anti-duplication -> No new alarms created
	for i := 6; i <= 7; i++ {
		healthN := &service.DeviceHealthInput{
			DeviceID:            dev.ID,
			State:               "ERROR",
			ConsecutiveFailures: i,
			LastErrorCategory:   "TIMEOUT",
			LastError:           errTimeout.Error(),
		}
		evaluator.EvaluateDeviceCommFailure(ctx, dev, healthN, errTimeout)
	}
	activeAfter, _ := alarmRepo.GetActive(ctx)
	if len(activeAfter) != 1 {
		t.Errorf("expected still only 1 active alarm after repeated failures, got %d", len(activeAfter))
	}

	// Recovery: Consecutive successes reach recovery limit -> Alarm automatically cleared
	healthSuccess := &service.DeviceHealthInput{
		DeviceID:             dev.ID,
		State:                "CONNECTED",
		ConsecutiveSuccesses: 2,
	}
	evaluator.EvaluateDeviceCommSuccess(ctx, dev, healthSuccess)

	if evaluator.GetActiveCommErrorsCount() != 0 {
		t.Errorf("expected comm error alarm cleared after recovery, got %d", evaluator.GetActiveCommErrorsCount())
	}
	clearedAlm, _ := alarmRepo.GetByID(ctx, active[0].ID)
	if clearedAlm.Status != model.AlarmCleared {
		t.Errorf("expected alarm status CLEARED, got %s", clearedAlm.Status)
	}
}

// 18. Test Communication Failure Categories
func TestAlarmEvaluator_CommErrorCategories(t *testing.T) {
	categories := []struct {
		errStr   string
		expected string
	}{
		{"connection reset by peer", "CONNECTION_REFUSED"},
		{"modbus crc exception checksum error", "PROTOCOL_EXCEPTION"},
		{"request timed out waiting for response", "TIMEOUT"},
		{"unexpected socket shutdown", "COMMUNICATION_FAILURE"},
	}

	for _, tc := range categories {
		db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
		ctx := context.Background()

		dev, _ := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)
		errComm := fmt.Errorf("%s", tc.errStr)

		health := &service.DeviceHealthInput{
			DeviceID:            dev.ID,
			ConsecutiveFailures: 5,
			LastErrorCategory:   tc.expected,
			LastError:           tc.errStr,
		}
		evaluator.EvaluateDeviceCommFailure(ctx, dev, health, errComm)

		active, err := alarmRepo.GetActive(ctx)
		if err != nil || len(active) != 1 {
			t.Fatalf("[%s] expected 1 active alarm, got %d", tc.expected, len(active))
		}
		if !hasSubstr(active[0].Message, tc.expected) {
			t.Errorf("[%s] expected message to contain category %s, got %s", tc.expected, tc.expected, active[0].Message)
		}
	}
}

// 19. Test Data Abnormality: Stuck Value Count, Duration, and Recovery
func TestAlarmEvaluator_DataAbnormalityStuckValue(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, param := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	startTime := time.Now().UTC().Add(-40 * time.Second)

	// Send 9 identical values -> Under count limit (10) -> No alarm
	for i := 0; i < 9; i++ {
		tStep := startTime.Add(time.Duration(i*3) * time.Second)
		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    param.ID,
			ParameterCode:  param.ParameterCode,
			ParameterName:  param.ParameterName,
			ProcessedValue: 42.50,
			Quality:        model.QualityGood,
			ReceivedAt:     tStep,
		}, param)
	}

	if evaluator.GetActiveAbnormalityCount() != 0 {
		t.Fatalf("expected 0 abnormality alarms under count limit, got %d", evaluator.GetActiveAbnormalityCount())
	}

	// 10th identical reading with duration > 30s -> Triggers DATA_ABNORMALITY_STUCK_VALUE
	t10 := startTime.Add(35 * time.Second)
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		ProcessedValue: 42.50,
		Quality:        model.QualityGood,
		ReceivedAt:     t10,
	}, param)

	if evaluator.GetActiveAbnormalityCount() != 1 {
		t.Fatalf("expected 1 active stuck value alarm, got %d", evaluator.GetActiveAbnormalityCount())
	}

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeAbnormalityStuck {
		t.Fatalf("expected active alarm code %s, got %+v", service.AlarmCodeAbnormalityStuck, active)
	}

	// Value changes to 43.10 -> Recovery! Stuck alarm is automatically cleared
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		ProcessedValue: 43.10, // Changed!
		Quality:        model.QualityGood,
		ReceivedAt:     t10.Add(1 * time.Second),
	}, param)

	if evaluator.GetActiveAbnormalityCount() != 0 {
		t.Errorf("expected stuck alarm cleared on value change, got %d", evaluator.GetActiveAbnormalityCount())
	}

	clearedAlm, _ := alarmRepo.GetByID(ctx, active[0].ID)
	if clearedAlm.Status != model.AlarmCleared {
		t.Errorf("expected alarm status CLEARED, got %s", clearedAlm.Status)
	}
}

// 20. Test Constant Environmental Readings Under Threshold Criteria Do Not False Alarm
func TestAlarmEvaluator_ConstantReadingSafeThresholds(t *testing.T) {
	db, evaluator, _, _ := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, param := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	// 5 identical samples over 10s (e.g. steady room temperature 24.0°C)
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    param.ID,
			ParameterCode:  param.ParameterCode,
			ParameterName:  param.ParameterName,
			ProcessedValue: 24.0,
			Quality:        model.QualityGood,
			ReceivedAt:     now.Add(time.Duration(i*2) * time.Second),
		}, param)
	}

	if evaluator.GetActiveCount() != 0 {
		t.Errorf("steady valid environmental readings must not trigger false alarms, got %d", evaluator.GetActiveCount())
	}
}

// 21. Test Repeated BAD Quality Sample Accumulation and Recovery
func TestAlarmEvaluator_DataAbnormalityBadQuality(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, param := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	// Samples 1-4: BAD quality -> Under limit of 5 -> No alarm
	for i := 1; i <= 4; i++ {
		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    param.ID,
			ParameterCode:  param.ParameterCode,
			ParameterName:  param.ParameterName,
			ProcessedValue: 0.0,
			Quality:        model.QualityBad,
			QualityReason:  model.ReasonDecodingError,
			ReceivedAt:     time.Now().UTC(),
		}, param)
	}
	if evaluator.GetActiveAbnormalityCount() != 0 {
		t.Fatalf("expected 0 alarms before 5th bad sample, got %d", evaluator.GetActiveAbnormalityCount())
	}

	// 5th BAD quality sample -> Triggers DATA_ABNORMALITY_BAD_QUALITY alarm
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		ProcessedValue: 0.0,
		Quality:        model.QualityBad,
		QualityReason:  model.ReasonDecodingError,
		ReceivedAt:     time.Now().UTC(),
	}, param)

	if evaluator.GetActiveAbnormalityCount() != 1 {
		t.Fatalf("expected 1 active bad quality alarm, got %d", evaluator.GetActiveAbnormalityCount())
	}

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeAbnormalityBad {
		t.Fatalf("expected active alarm code %s, got %+v", service.AlarmCodeAbnormalityBad, active)
	}

	// Healthy GOOD quality sample arrives -> Alarm cleared automatically
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		ProcessedValue: 55.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}, param)

	if evaluator.GetActiveAbnormalityCount() != 0 {
		t.Errorf("expected bad quality alarm cleared on Good sample, got %d", evaluator.GetActiveAbnormalityCount())
	}
	clearedAlm, _ := alarmRepo.GetByID(ctx, active[0].ID)
	if clearedAlm.Status != model.AlarmCleared {
		t.Errorf("expected alarm status CLEARED, got %s", clearedAlm.Status)
	}
}

// 22. Test Repeated Stale Data Abnormality and Recovery
func TestAlarmEvaluator_DataAbnormalityStaleData(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, param := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	// 3 consecutive STALE samples -> Triggers DATA_ABNORMALITY_STALE_DATA
	for i := 0; i < 3; i++ {
		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    param.ID,
			ParameterCode:  param.ParameterCode,
			ParameterName:  param.ParameterName,
			ProcessedValue: 50.0,
			Quality:        model.QualityStale,
			QualityFlags:   model.FlagStale,
			ReceivedAt:     time.Now().UTC(),
		}, param)
	}

	if evaluator.GetActiveAbnormalityCount() != 1 {
		t.Fatalf("expected 1 active stale data alarm, got %d", evaluator.GetActiveAbnormalityCount())
	}

	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeAbnormalityStale {
		t.Fatalf("expected active alarm code %s, got %+v", service.AlarmCodeAbnormalityStale, active)
	}

	// Good quality sample arrives -> Alarm cleared automatically
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		ProcessedValue: 50.5,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}, param)

	if evaluator.GetActiveAbnormalityCount() != 0 {
		t.Errorf("expected stale data alarm cleared on Good sample, got %d", evaluator.GetActiveAbnormalityCount())
	}
}

// 23. Test Spike / Outlier Abnormality and Recovery
func TestAlarmEvaluator_DataAbnormalitySpike(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, param := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	// Spike detected sample
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ParameterName:  param.ParameterName,
		ProcessedValue: 999.0,
		Quality:        model.QualityUncertain,
		QualityReason:  model.ReasonSpikeDetected,
		QualityFlags:   model.FlagSpike,
		ReceivedAt:     time.Now().UTC(),
	}, param)

	if evaluator.GetActiveAbnormalityCount() != 1 {
		t.Fatalf("expected 1 active spike alarm, got %d", evaluator.GetActiveAbnormalityCount())
	}

	// 3 normal samples -> Spike alarm clears
	for i := 0; i < 3; i++ {
		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    param.ID,
			ParameterCode:  param.ParameterCode,
			ParameterName:  param.ParameterName,
			ProcessedValue: 50.0 + float64(i),
			Quality:        model.QualityGood,
			ReceivedAt:     time.Now().UTC(),
		}, param)
	}

	if evaluator.GetActiveAbnormalityCount() != 0 {
		t.Errorf("expected spike alarm cleared after 3 normal readings, got %d", evaluator.GetActiveAbnormalityCount())
	}
	active, _ := alarmRepo.GetActive(ctx)
	if len(active) != 0 {
		t.Errorf("expected 0 active alarms in DB, got %d", len(active))
	}
}

// 24. Test Restart Reconciliation for Wave 2 Alarms
func TestAlarmEvaluator_RestartReconciliationWave2(t *testing.T) {
	db, _, alarmService, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, param := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	// Seed 3 alarms into DB: Device Offline, Comm Error, Data Abnormality Stuck
	alarmOffline := &model.Alarm{
		DeviceID:    &dev.ID,
		AlarmCode:   service.AlarmCodeDeviceOffline,
		Severity:    model.SeverityCritical,
		Message:     "Device offline outage",
		Status:      model.AlarmActive,
		TriggeredAt: time.Now().UTC(),
	}
	_ = alarmService.CreateAlarm(ctx, alarmOffline)

	alarmComm := &model.Alarm{
		DeviceID:    &dev.ID,
		AlarmCode:   service.AlarmCodeCommError,
		Severity:    model.SeverityWarning,
		Message:     "Comm timeout",
		Status:      model.AlarmActive,
		TriggeredAt: time.Now().UTC(),
	}
	_ = alarmService.CreateAlarm(ctx, alarmComm)

	alarmStuck := &model.Alarm{
		DeviceID:    &dev.ID,
		ParameterID: &param.ID,
		AlarmCode:   service.AlarmCodeAbnormalityStuck,
		Severity:    model.SeverityWarning,
		Message:     "Stuck reading",
		Status:      model.AlarmActive,
		TriggeredAt: time.Now().UTC(),
	}
	_ = alarmService.CreateAlarm(ctx, alarmStuck)

	// Simulate system restart: instantiate fresh evaluator
	cfg := service.DefaultAlarmEvaluatorConfig()
	freshEvaluator := service.NewAlarmEvaluator(ctx, alarmService, cfg, nil)

	if freshEvaluator.GetActiveDeviceOfflineCount() != 1 {
		t.Errorf("expected 1 offline alarm reconciled from DB, got %d", freshEvaluator.GetActiveDeviceOfflineCount())
	}
	if freshEvaluator.GetActiveCommErrorsCount() != 1 {
		t.Errorf("expected 1 comm error alarm reconciled from DB, got %d", freshEvaluator.GetActiveCommErrorsCount())
	}
	if freshEvaluator.GetActiveAbnormalityCount() != 1 {
		t.Errorf("expected 1 abnormality alarm reconciled from DB, got %d", freshEvaluator.GetActiveAbnormalityCount())
	}

	// Device reconnects and communicates successfully -> Both offline and comm error alarms resolved
	now := time.Now().UTC()
	freshEvaluator.EvaluateDeviceCommSuccess(ctx, dev, &service.DeviceHealthInput{
		DeviceID:             dev.ID,
		State:                "CONNECTED",
		ConsecutiveSuccesses: 3,
		LastCommSuccess:      &now,
	})

	if freshEvaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Errorf("expected offline alarm cleared on recovery after restart, got %d", freshEvaluator.GetActiveDeviceOfflineCount())
	}
	if freshEvaluator.GetActiveCommErrorsCount() != 0 {
		t.Errorf("expected comm error alarm cleared on recovery after restart, got %d", freshEvaluator.GetActiveCommErrorsCount())
	}

	clearedOffline, _ := alarmRepo.GetByID(ctx, alarmOffline.ID)
	if clearedOffline.Status != model.AlarmCleared {
		t.Errorf("expected offline alarm status CLEARED, got %s", clearedOffline.Status)
	}
}

// 25. Test Concurrent Wave 2 Safety and Race Protection
func TestAlarmEvaluator_ConcurrentWave2Safety(t *testing.T) {
	db, evaluator, _, _ := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, param := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	var wg sync.WaitGroup
	iterations := 100

	// Goroutine 1: Rapid telemetry evaluations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
				DeviceID:       dev.ID,
				DeviceCode:     dev.DeviceCode,
				ParameterID:    param.ID,
				ParameterCode:  param.ParameterCode,
				ParameterName:  param.ParameterName,
				ProcessedValue: 50.0 + float64(i%3),
				Quality:        model.QualityGood,
				ReceivedAt:     time.Now().UTC(),
			}, param)
		}
	}()

	// Goroutine 2: Communication errors and successes
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			if i%5 == 0 {
				evaluator.EvaluateDeviceCommSuccess(ctx, dev, &service.DeviceHealthInput{
					DeviceID:             dev.ID,
					State:                "CONNECTED",
					ConsecutiveSuccesses: 2,
				})
			} else {
				evaluator.EvaluateDeviceCommFailure(ctx, dev, &service.DeviceHealthInput{
					DeviceID:            dev.ID,
					State:               "ERROR",
					ConsecutiveFailures: i,
				}, fmt.Errorf("transient comm error"))
			}
		}
	}()

	// Goroutine 3: Device offline evaluations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			tRecent := time.Now().UTC().Add(-time.Duration(i) * time.Second)
			evaluator.EvaluateDeviceOffline(ctx, dev, &service.DeviceHealthInput{
				DeviceID:        dev.ID,
				State:           "CONNECTED",
				LastCommSuccess: &tRecent,
			})
		}
	}()

	wg.Wait()

	// Metrics sanity check
	evals, _, _, _ := evaluator.GetMetrics()
	if evals == 0 {
		t.Errorf("expected >0 evaluations tracked")
	}
}

// 29. Regression Test: Actively communicating device must never trigger false DEVICE_OFFLINE alarms (0s outage flapping)
func TestAlarmEvaluator_NoFalseDeviceOfflineOnRecentCommunication(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	dev, _ := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)

	// Step 1: Device communication just succeeded (0s outage, consecutive successes = 1)
	now := time.Now().UTC()
	health := &service.DeviceHealthInput{
		DeviceID:             dev.ID,
		State:                "CONNECTED",
		ConsecutiveSuccesses: 1,
		ConsecutiveFailures:  0,
		LastCommSuccess:      &now,
		LastErrorCategory:    "",
		LastError:            "",
	}

	evaluator.EvaluateDeviceOffline(ctx, dev, health)
	if evaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Fatalf("expected 0 offline alarms on actively communicating device, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	// Step 2: Even if an old error state lingers or was just cleared, positive successes + recent comm must not trigger offline
	healthWithResidualErrorState := &service.DeviceHealthInput{
		DeviceID:             dev.ID,
		State:                "ERROR",
		ConsecutiveSuccesses: 1,
		ConsecutiveFailures:  0,
		LastCommSuccess:      &now,
		LastErrorCategory:    "CONNECTION_REFUSED",
	}

	evaluator.EvaluateDeviceOffline(ctx, dev, healthWithResidualErrorState)
	if evaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Fatalf("expected 0 offline alarms when device has positive successes and recent comm, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	// Step 3: Simulate recovery cycle: previous offline alarm is cleared by CommSuccess and not re-triggered by Offline evaluation
	// Force an offline alarm first (> 60s timeout)
	oldComm := time.Now().UTC().Add(-100 * time.Second)
	offlineHealth := &service.DeviceHealthInput{
		DeviceID:             dev.ID,
		State:                "ERROR",
		ConsecutiveSuccesses: 0,
		ConsecutiveFailures:  5,
		LastCommSuccess:      &oldComm,
		LastErrorCategory:    "TIMEOUT",
	}
	evaluator.EvaluateDeviceOffline(ctx, dev, offlineHealth)
	if evaluator.GetActiveDeviceOfflineCount() != 1 {
		t.Fatalf("expected 1 offline alarm during genuine timeout, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	// Now successful poll occurs:
	recoveredTime := time.Now().UTC()
	recoveredHealth := &service.DeviceHealthInput{
		DeviceID:             dev.ID,
		State:                "CONNECTED",
		ConsecutiveSuccesses: 2,
		ConsecutiveFailures:  0,
		LastCommSuccess:      &recoveredTime,
	}
	evaluator.EvaluateDeviceCommSuccess(ctx, dev, recoveredHealth)
	if evaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Fatalf("expected 0 offline alarms after CommSuccess recovery, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	// Immediately evaluate offline check at end of poll cycle — must stay cleared, no flapping!
	evaluator.EvaluateDeviceOffline(ctx, dev, recoveredHealth)
	if evaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Fatalf("flapping detected: offline alarm was falsely re-triggered on recovered device! got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	active, err := alarmRepo.GetActive(ctx)
	if err != nil || len(active) != 0 {
		t.Fatalf("expected 0 active alarms in repository, got %d (err: %v)", len(active), err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && hasSubstr(s, substr)))
}

func hasSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestAlarmEvaluator_ZeroValueAlarmPolicies(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	cLow := 5.0
	wLow := 10.0

	// 1. SUPPRESS_THRESHOLD Policy (e.g. Wind Speed, Rainfall, Solar Radiation)
	dev, paramSuppress := createTestDeviceAndParam(db, &wLow, nil, &cLow, nil, nil)
	paramSuppress.ZeroValueAlarmPolicy = model.ZeroValuePolicySuppressThreshold
	db.Save(paramSuppress)

	// Value 0.0 must NOT trigger threshold alarms
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    paramSuppress.ID,
		ParameterCode:  paramSuppress.ParameterCode,
		ParameterName:  paramSuppress.ParameterName,
		ProcessedValue: 0.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}, paramSuppress)

	active, err := alarmRepo.GetActive(ctx)
	if err != nil {
		t.Fatalf("failed to query active alarms: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("expected 0 alarms for SUPPRESS_THRESHOLD with value 0.0, got %d (code=%s)", len(active), active[0].AlarmCode)
	}

	// Non-zero threshold breach (3.0 < 5.0) MUST trigger threshold alarm
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    paramSuppress.ID,
		ParameterCode:  paramSuppress.ParameterCode,
		ParameterName:  paramSuppress.ParameterName,
		ProcessedValue: 3.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}, paramSuppress)

	active, _ = alarmRepo.GetActive(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeCriticalLow {
		t.Fatalf("expected 1 CriticalLow alarm for value 3.0, got %d", len(active))
	}

	// Returning to 0.0 under SUPPRESS_THRESHOLD must clear the threshold alarm
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    paramSuppress.ID,
		ParameterCode:  paramSuppress.ParameterCode,
		ParameterName:  paramSuppress.ParameterName,
		ProcessedValue: 0.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}, paramSuppress)

	active, _ = alarmRepo.GetActive(ctx)
	if len(active) != 0 {
		t.Fatalf("expected threshold alarm to clear on return to 0.0 under SUPPRESS_THRESHOLD, got %d", len(active))
	}

	// 2. NORMAL Policy (Default behavior)
	devNormal, paramNormal := createTestDeviceAndParam(db, &wLow, nil, &cLow, nil, nil)
	paramNormal.ZeroValueAlarmPolicy = model.ZeroValuePolicyNormal
	db.Save(paramNormal)

	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       devNormal.ID,
		DeviceCode:     devNormal.DeviceCode,
		ParameterID:    paramNormal.ID,
		ParameterCode:  paramNormal.ParameterCode,
		ParameterName:  paramNormal.ParameterName,
		ProcessedValue: 0.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}, paramNormal)

	active, _ = alarmRepo.GetActive(ctx)
	if len(active) != 1 || active[0].AlarmCode != service.AlarmCodeCriticalLow {
		t.Fatalf("expected CriticalLow alarm for NORMAL policy with value 0.0, got %d", len(active))
	}

	// 3. INVALID_READING Policy
	devInvalid, paramInvalid := createTestDeviceAndParam(db, nil, nil, nil, nil, nil)
	paramInvalid.ZeroValueAlarmPolicy = model.ZeroValuePolicyInvalidReading
	db.Save(paramInvalid)

	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       devInvalid.ID,
		DeviceCode:     devInvalid.DeviceCode,
		ParameterID:    paramInvalid.ID,
		ParameterCode:  paramInvalid.ParameterCode,
		ParameterName:  paramInvalid.ParameterName,
		ProcessedValue: 0.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}, paramInvalid)

	active, _ = alarmRepo.GetActive(ctx)
	foundInvalid := false
	for _, a := range active {
		if a.AlarmCode == service.AlarmCodeAbnormalityZeroInvalid {
			foundInvalid = true
			break
		}
	}
	if !foundInvalid {
		t.Fatalf("expected DATA_ABNORMALITY_ZERO_INVALID alarm for INVALID_READING policy, got %v", active)
	}

	// Returning to non-zero (5.0) clears the invalid alarm
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       devInvalid.ID,
		DeviceCode:     devInvalid.DeviceCode,
		ParameterID:    paramInvalid.ID,
		ParameterCode:  paramInvalid.ParameterCode,
		ParameterName:  paramInvalid.ParameterName,
		ProcessedValue: 5.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}, paramInvalid)

	active, _ = alarmRepo.GetActive(ctx)
	for _, a := range active {
		if a.AlarmCode == service.AlarmCodeAbnormalityZeroInvalid {
			t.Fatalf("expected DATA_ABNORMALITY_ZERO_INVALID to clear for non-zero value")
		}
	}

	// 4. Device Offline is NOT suppressed by zero policy
	dev.Enabled = true
	dev.Status = model.DeviceStatusActive
	oldSuccess := time.Now().UTC().Add(-75 * time.Second)
	devHealth := &service.DeviceHealthInput{
		DeviceID:            dev.ID,
		State:               "ERROR",
		ConsecutiveFailures: 10,
		LastCommSuccess:     &oldSuccess,
		LastErrorCategory:   "TIMEOUT",
		LastError:           "i/o timeout",
	}
	evaluator.EvaluateDeviceOffline(ctx, dev, devHealth)
	active, _ = alarmRepo.GetActive(ctx)
	foundOffline := false
	for _, a := range active {
		if a.AlarmCode == service.AlarmCodeDeviceOffline {
			foundOffline = true
			break
		}
	}
	if !foundOffline {
		t.Fatalf("expected DEVICE_OFFLINE alarm to trigger regardless of zero value policy")
	}
}

func TestAlarmEvaluator_DisabledDeviceClearsAllAlarms(t *testing.T) {
	db, evaluator, _, alarmRepo := setupTestEvaluatorEnvironment(t)
	ctx := context.Background()

	critHigh := 100.0
	dev, param := createTestDeviceAndParam(db, nil, nil, nil, &critHigh, nil)

	// Set deviceLookup
	evaluator.SetDeviceLookup(func(ctx context.Context, deviceID uint) (*model.Device, error) {
		var d model.Device
		if err := db.First(&d, deviceID).Error; err != nil {
			return nil, err
		}
		return &d, nil
	})

	// 1. Trigger threshold alarm
	tHigh := &service.LatestTelemetry{
		DeviceID:       dev.ID,
		DeviceCode:     dev.DeviceCode,
		ParameterID:    param.ID,
		ParameterCode:  param.ParameterCode,
		ProcessedValue: 120.0,
		Quality:        model.QualityGood,
		ReceivedAt:     time.Now().UTC(),
	}
	evaluator.EvaluateTelemetry(ctx, tHigh, param)
	if evaluator.GetActiveCount() != 1 {
		t.Fatalf("expected 1 active threshold alarm, got %d", evaluator.GetActiveCount())
	}

	// 2. Trigger offline alarm
	oldSuccess := time.Now().UTC().Add(-120 * time.Second)
	health := &service.DeviceHealthInput{
		DeviceID:        dev.ID,
		State:           "ERROR",
		LastCommSuccess: &oldSuccess,
	}
	evaluator.EvaluateDeviceOffline(ctx, dev, health)
	if evaluator.GetActiveDeviceOfflineCount() != 1 {
		t.Fatalf("expected 1 active offline alarm, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	// 3. Disable the device in DB
	db.Model(&model.Device{}).Where("id = ?", dev.ID).Update("enabled", false)

	// 4. Call ClearDeviceAlarms
	evaluator.ClearDeviceAlarms(ctx, dev.ID)

	if evaluator.GetActiveCount() != 0 {
		t.Errorf("expected 0 active threshold alarms after ClearDeviceAlarms, got %d", evaluator.GetActiveCount())
	}
	if evaluator.GetActiveDeviceOfflineCount() != 0 {
		t.Errorf("expected 0 active offline alarms after ClearDeviceAlarms, got %d", evaluator.GetActiveDeviceOfflineCount())
	}

	// Verify DB active alarms are cleared
	active, err := alarmRepo.GetActive(ctx)
	if err != nil {
		t.Fatalf("failed to query active alarms: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("expected 0 active alarms in DB for disabled device, got %d", len(active))
	}
}



