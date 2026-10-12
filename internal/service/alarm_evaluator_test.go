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
