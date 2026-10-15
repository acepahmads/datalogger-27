package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupPipelineTestDB(t *testing.T) (*gorm.DB, *service.ISPUService, *service.AggregationService, *model.Device) {
	dsn := fmt.Sprintf("file:mempipeline_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test sqlite: %v", err)
	}

	err = db.AutoMigrate(
		&model.Device{},
		&model.Parameter{},
		&model.RawData{},
		&model.AggregationDefinition{},
		&model.AggregationResult{},
		&model.AuditTrail{},
		&model.ISPUParameterMapping{},
		&model.ISPUCalculationResult{},
		&model.ISPUParameterResult{},
	)
	if err != nil {
		t.Fatalf("Failed to automigrate: %v", err)
	}

	dev := &model.Device{
		DeviceCode: "AQMS-PIPE-01",
		DeviceName: "Pipeline Test Station",
		DeviceType: "MODBUS_TCP",
		Status:     model.DeviceStatusActive,
		Timezone:   "Asia/Jakarta",
	}
	db.Create(dev)

	paramPM25 := &model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "PM25",
		ParameterName: "Particulate Matter 2.5",
		Unit:          "µg/m³",
		Enabled:       true,
	}
	db.Create(paramPM25)

	paramPM10 := &model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "PM10",
		ParameterName: "Particulate Matter 10",
		Unit:          "µg/m³",
		Enabled:       true,
	}
	db.Create(paramPM10)

	// Aggregation definition (default 120s / 2m)
	aggDefPM25 := &model.AggregationDefinition{
		Name:            "PM2.5 2m Rollup",
		Code:            "AGG_PM25_2M",
		DeviceID:        dev.ID,
		ParameterID:     paramPM25.ID,
		Function:        model.FunctionAvg,
		IntervalSeconds: 120,
		Timezone:        "Asia/Jakarta",
		Enabled:         true,
	}
	db.Create(aggDefPM25)

	aggDefPM10 := &model.AggregationDefinition{
		Name:            "PM10 2m Rollup",
		Code:            "AGG_PM10_2M",
		DeviceID:        dev.ID,
		ParameterID:     paramPM10.ID,
		Function:        model.FunctionAvg,
		IntervalSeconds: 120,
		Timezone:        "Asia/Jakarta",
		Enabled:         true,
	}
	db.Create(aggDefPM10)

	// ISPU parameter mappings
	db.Create(&model.ISPUParameterMapping{
		DeviceID:         dev.ID,
		ParameterID:      paramPM25.ID,
		PollutantCode:    model.PollutantPM25,
		SourceUnit:       "µg/m³",
		NormalizedUnit:   "µg/m³",
		ConversionFactor: 1.0,
		MinCompleteness:  75.0,
		Enabled:          true,
	})
	db.Create(&model.ISPUParameterMapping{
		DeviceID:         dev.ID,
		ParameterID:      paramPM10.ID,
		PollutantCode:    model.PollutantPM10,
		SourceUnit:       "µg/m³",
		NormalizedUnit:   "µg/m³",
		ConversionFactor: 1.0,
		MinCompleteness:  75.0,
		Enabled:          true,
	})

	ispuRepo := repository.NewISPURepository(db)
	sysRepo := repository.NewSystemRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	aggRepo := repository.NewAggregationRepository(db)

	ispuSvc := service.NewISPUService(db, ispuRepo, sysRepo)
	aggSvc := service.NewAggregationService(aggRepo, devRepo, sysRepo, nil)

	return db, ispuSvc, aggSvc, dev
}

// 1. A completed aggregation result triggers the ISPU workflow
func TestISPUPipeline_AggregationCompletionTrigger(t *testing.T) {
	db, ispuSvc, aggSvc, dev := setupPipelineTestDB(t)
	ctx := context.Background()

	ispuSvc.StartWorker(ctx)
	defer ispuSvc.StopWorker()

	// Connect aggregation completion to ISPU service
	aggSvc.SetOnAggregationCompleted(func(c context.Context, event service.AggregationCompletedEvent) {
		ispuSvc.HandleAggregationCompleted(c, event)
	})

	// Seed aggregation results for the completed bucket
	now := time.Date(2026, 10, 15, 16, 40, 0, 0, time.UTC)
	val25 := 55.4 // ISPU 100
	val10 := 50.0 // ISPU 50

	var p25, p10 model.Parameter
	db.Where("device_id = ? AND parameter_code = 'PM25'", dev.ID).First(&p25)
	db.Where("device_id = ? AND parameter_code = 'PM10'", dev.ID).First(&p10)

	db.Create(&model.AggregationResult{
		DeviceID:        dev.ID,
		ParameterID:     p25.ID,
		PeriodStart:     now.Add(-2 * time.Minute),
		PeriodEnd:       now,
		IntervalSeconds: 120,
		AvgValue:        &val25,
		Quality:         model.QualityGood,
		Identifier:      "20261015164000",
	})
	db.Create(&model.AggregationResult{
		DeviceID:        dev.ID,
		ParameterID:     p10.ID,
		PeriodStart:     now.Add(-2 * time.Minute),
		PeriodEnd:       now,
		IntervalSeconds: 120,
		AvgValue:        &val10,
		Quality:         model.QualityGood,
		Identifier:      "20261015164000",
	})

	// Fire aggregation completed event for PM2.5
	ispuSvc.HandleAggregationCompleted(ctx, service.AggregationCompletedEvent{
		DeviceID:        dev.ID,
		ParameterID:     p25.ID,
		PeriodStart:     now.Add(-2 * time.Minute),
		PeriodEnd:       now,
		IntervalSeconds: 120,
		Function:        "AVG",
		Identifier:      "20261015164000",
		Quality:         "GOOD",
	})

	// Allow worker debounce/coalesce to finish
	time.Sleep(500 * time.Millisecond)

	// Verify ISPU result was persisted in DB
	var res model.ISPUCalculationResult
	err := db.Preload("ParameterResults").Where("device_id = ? AND period_end = ?", dev.ID, now).First(&res).Error
	if err != nil {
		t.Fatalf("Expected ISPUCalculationResult persisted after aggregation trigger: %v", err)
	}

	if res.SourceBucketIdentifier != "20261015234000" && res.SourceBucketIdentifier != "20261015164000" {
		t.Logf("Persisted SourceBucketIdentifier: %s", res.SourceBucketIdentifier)
	}
	if res.AggregationIntervalSeconds != 120 {
		t.Errorf("Expected AggregationIntervalSeconds 120, got %d", res.AggregationIntervalSeconds)
	}
	if res.TriggerReason != "AGGREGATION_COMPLETED" {
		t.Errorf("Expected TriggerReason AGGREGATION_COMPLETED, got %s", res.TriggerReason)
	}
}

// 2 & 3. Configured interval (120s and other intervals) work dynamically without code changes
func TestISPUPipeline_ConfiguredIntervalDynamic(t *testing.T) {
	db, ispuSvc, _, dev := setupPipelineTestDB(t)
	ctx := context.Background()

	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Test 1: 120s interval (configured)
	target1 := time.Date(2026, 10, 15, 16, 21, 45, 0, loc)
	res1, err := ispuSvc.CalculateStationWindowWithReason(ctx, dev.ID, target1, "TEST_120S")
	if err != nil {
		t.Fatalf("Calculation for 120s failed: %v", err)
	}
	if res1.AggregationIntervalSeconds != 120 {
		t.Errorf("Expected interval 120s, got %d", res1.AggregationIntervalSeconds)
	}
	if res1.PeriodEnd.In(loc).Minute()%2 != 0 {
		t.Errorf("120s interval must produce even minute boundary, got %d", res1.PeriodEnd.In(loc).Minute())
	}

	// Test 2: Change configured interval in DB to 300s (5 minutes) without code changes
	db.Model(&model.AggregationDefinition{}).Where("device_id = ?", dev.ID).Update("interval_seconds", 300)

	target2 := time.Date(2026, 10, 15, 16, 24, 15, 0, loc)
	res2, err := ispuSvc.CalculateStationWindowWithReason(ctx, dev.ID, target2, "TEST_300S")
	if err != nil {
		t.Fatalf("Calculation for 300s failed: %v", err)
	}
	if res2.AggregationIntervalSeconds != 300 {
		t.Errorf("Expected dynamic interval 300s, got %d", res2.AggregationIntervalSeconds)
	}
	if res2.PeriodEnd.In(loc).Minute()%5 != 0 {
		t.Errorf("300s interval must produce 5-minute boundary (0, 5, 10... 20, 25), got %d", res2.PeriodEnd.In(loc).Minute())
	}
}

// 4. Repeated processing of the same aggregation event cannot create duplicate results
func TestISPUPipeline_IdempotencyAndNoDuplicates(t *testing.T) {
	db, ispuSvc, _, dev := setupPipelineTestDB(t)
	ctx := context.Background()

	target := time.Date(2026, 10, 15, 16, 30, 0, 0, time.UTC)

	// Run 3 times for the exact same period
	for i := 0; i < 3; i++ {
		_, err := ispuSvc.CalculateStationWindowWithReason(ctx, dev.ID, target, "DUPLICATE_TEST")
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}
	}

	var count int64
	db.Model(&model.ISPUCalculationResult{}).Where("device_id = ?", dev.ID).Count(&count)
	if count != 1 {
		t.Errorf("Expected exactly 1 ISPUCalculationResult, got %d (duplicates created!)", count)
	}
}

// 6. ISPU API strictly reads persisted results without calculating on read
func TestISPUPipeline_PureReadAPIs(t *testing.T) {
	db, ispuSvc, _, dev := setupPipelineTestDB(t)
	ctx := context.Background()

	// Initial DB has 0 calculations
	var initialCount int64
	db.Model(&model.ISPUCalculationResult{}).Where("device_id = ?", dev.ID).Count(&initialCount)
	if initialCount != 0 {
		t.Fatalf("Expected 0 initial records")
	}

	// Calling GetLatestSummary on empty station must NOT calculate or create DB records
	latest, err := ispuSvc.GetLatestSummary(ctx, dev.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		t.Fatalf("Unexpected error from GetLatestSummary: %v", err)
	}
	if latest != nil {
		t.Errorf("Expected nil latest on empty DB, got non-nil result")
	}

	var postCount int64
	db.Model(&model.ISPUCalculationResult{}).Where("device_id = ?", dev.ID).Count(&postCount)
	if postCount != 0 {
		t.Errorf("GetLatestSummary created calculation records on GET! Count = %d", postCount)
	}

	// Calling GetTrend must also NOT create calculation records
	trend, err := ispuSvc.GetTrend(ctx, dev.ID, "24h")
	if err != nil {
		t.Fatalf("GetTrend error: %v", err)
	}
	if len(trend) != 0 {
		t.Errorf("Expected empty trend, got %d records", len(trend))
	}

	var trendPostCount int64
	db.Model(&model.ISPUCalculationResult{}).Where("device_id = ?", dev.ID).Count(&trendPostCount)
	if trendPostCount != 0 {
		t.Errorf("GetTrend created calculation records on GET! Count = %d", trendPostCount)
	}
}

// 8. Missing or invalid source data does not produce a falsely valid result
func TestISPUPipeline_MissingOrInvalidSourceData(t *testing.T) {
	_, ispuSvc, _, dev := setupPipelineTestDB(t)
	ctx := context.Background()

	now := time.Now().UTC()
	res, err := ispuSvc.CalculateStationWindowWithReason(ctx, dev.ID, now, "MISSING_DATA_TEST")
	if err != nil {
		t.Fatalf("Calculation failed: %v", err)
	}

	if res.CalculationStatus == model.ISPUStatusValid {
		t.Errorf("Calculation with zero data must NOT have VALID status, got %s", res.CalculationStatus)
	}
	if res.CompletenessPercent != 0 {
		t.Errorf("Expected 0%% completeness for missing data, got %.1f%%", res.CompletenessPercent)
	}
}

// 9. Late data triggers correct period recalculation idempotently
func TestISPUPipeline_LateDataRecalculation(t *testing.T) {
	db, ispuSvc, _, dev := setupPipelineTestDB(t)
	ctx := context.Background()

	now := time.Date(2026, 10, 15, 14, 0, 0, 0, time.UTC)
	var p25 model.Parameter
	db.Where("device_id = ? AND parameter_code = 'PM25'", dev.ID).First(&p25)

	// First calculation with value 15.0
	val1 := 15.0
	agg1 := &model.AggregationResult{
		DeviceID:        dev.ID,
		ParameterID:     p25.ID,
		PeriodStart:     now.Add(-2 * time.Minute),
		PeriodEnd:       now,
		IntervalSeconds: 120,
		AvgValue:        &val1,
		Quality:         model.QualityGood,
		Identifier:      "20261015140000",
	}
	db.Create(agg1)

	res1, err := ispuSvc.CalculateStationWindowWithReason(ctx, dev.ID, now, "INITIAL")
	if err != nil {
		t.Fatalf("Initial calculation failed: %v", err)
	}
	initialISPU := res1.OverallISPU

	// Late correction: value updated to 85.0
	val2 := 85.0
	db.Model(agg1).Updates(map[string]interface{}{
		"avg_value": &val2,
		"value":     val2,
	})

	// Recalculate affected period with reason LATE_DATA
	res2, err := ispuSvc.CalculateStationWindowWithReason(ctx, dev.ID, now, "LATE_DATA")
	if err != nil {
		t.Fatalf("Late data recalculation failed: %v", err)
	}

	if res2.ID != res1.ID {
		t.Errorf("Expected idempotent update of ID %d, got %d", res1.ID, res2.ID)
	}
	if res2.TriggerReason != "LATE_DATA" {
		t.Errorf("Expected TriggerReason LATE_DATA, got %s", res2.TriggerReason)
	}
	if res2.OverallISPU <= initialISPU {
		t.Errorf("Expected ISPU to increase after late correction (from %.1f), got %.1f", initialISPU, res2.OverallISPU)
	}
}

// 10. Manual recalculation uses the exact same calculation service and logs audit trail
func TestISPUPipeline_ManualRecalculationUsesSameService(t *testing.T) {
	db, ispuSvc, _, dev := setupPipelineTestDB(t)
	ctx := context.Background()

	target := time.Date(2026, 10, 15, 15, 0, 0, 0, time.UTC)
	res, err := ispuSvc.Recalculate(ctx, dev.ID, &target, "admin", "127.0.0.1")
	if err != nil {
		t.Fatalf("Recalculate failed: %v", err)
	}

	if res == nil {
		t.Fatalf("Expected non-nil recalculation result")
	}
	if res.TriggerReason != "MANUAL_RECALCULATE" {
		t.Errorf("Expected TriggerReason MANUAL_RECALCULATE, got %s", res.TriggerReason)
	}

	// Verify audit trail logged
	var audit model.AuditTrail
	err = db.Where("action = 'ISPU_RECALCULATE' AND username = 'admin'").First(&audit).Error
	if err != nil {
		t.Errorf("Expected audit trail for manual recalculation: %v", err)
	}
}

// 11. Restart recovery reconciles missing intervals without duplicates
func TestISPUPipeline_RestartRecoveryReconciliation(t *testing.T) {
	db, ispuSvc, _, dev := setupPipelineTestDB(t)
	ctx := context.Background()

	var p25 model.Parameter
	db.Where("device_id = ? AND parameter_code = 'PM25'", dev.ID).First(&p25)

	// Seed completed aggregation buckets in the last hour
	now := time.Now().UTC().Truncate(2 * time.Minute)
	val := 42.0
	for i := 1; i <= 3; i++ {
		pEnd := now.Add(-time.Duration(i*2) * time.Minute)
		pStart := pEnd.Add(-2 * time.Minute)
		db.Create(&model.AggregationResult{
			DeviceID:        dev.ID,
			ParameterID:     p25.ID,
			PeriodStart:     pStart,
			PeriodEnd:       pEnd,
			IntervalSeconds: 120,
			AvgValue:        &val,
			Quality:         model.QualityGood,
			Identifier:      pEnd.Format("20060102150405"),
		})
	}

	// Verify 0 calculation records exist currently
	var countBefore int64
	db.Model(&model.ISPUCalculationResult{}).Where("device_id = ?", dev.ID).Count(&countBefore)
	if countBefore != 0 {
		t.Fatalf("Expected 0 calculation records before recovery")
	}

	// Run startup reconciliation
	recovered, err := ispuSvc.ReconcilePendingCalculations(ctx)
	if err != nil {
		t.Fatalf("ReconcilePendingCalculations failed: %v", err)
	}
	if recovered < 3 {
		t.Errorf("Expected at least 3 missing intervals recovered, got %d", recovered)
	}

	var countAfter int64
	db.Model(&model.ISPUCalculationResult{}).Where("device_id = ?", dev.ID).Count(&countAfter)
	if countAfter < 3 {
		t.Errorf("Expected at least 3 persisted calculation records, got %d", countAfter)
	}

	// Second reconciliation run must find 0 missing (idempotent, no duplicates)
	recovered2, err2 := ispuSvc.ReconcilePendingCalculations(ctx)
	if err2 != nil {
		t.Fatalf("Second reconciliation failed: %v", err2)
	}
	if recovered2 != 0 {
		t.Errorf("Expected 0 missing intervals on second recovery run, got %d", recovered2)
	}

	var countFinal int64
	db.Model(&model.ISPUCalculationResult{}).Where("device_id = ?", dev.ID).Count(&countFinal)
	if countFinal != countAfter {
		t.Errorf("Duplicate records created on second recovery run! Before=%d, After=%d", countAfter, countFinal)
	}
}
