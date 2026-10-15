package service_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupISPUTestDB(t *testing.T) (*gorm.DB, *service.ISPUService, *model.Device) {
	dsn := fmt.Sprintf("file:memispu_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
		&model.AggregationResult{},
		&model.AuditTrail{},
		&model.ISPUParameterMapping{},
		&model.ISPUCalculationResult{},
		&model.ISPUParameterResult{},
	)
	if err != nil {
		t.Fatalf("Failed to automigrate: %v", err)
	}

	// Create test AQMS Device
	dev := &model.Device{
		DeviceCode: "AQMS-TEST-01",
		DeviceName: "AQMS Stasiun Pengujian",
		DeviceType: "MODBUS_TCP",
		Status:     model.DeviceStatusActive,
		Timezone:   "Asia/Jakarta",
	}
	db.Create(dev)

	// Create 2 test parameters: PM2.5 and PM10
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

	// Create mappings
	mappingPM25 := &model.ISPUParameterMapping{
		DeviceID:         dev.ID,
		ParameterID:      paramPM25.ID,
		PollutantCode:    model.PollutantPM25,
		SourceUnit:       "µg/m³",
		NormalizedUnit:   "µg/m³",
		ConversionFactor: 1.0,
		MinCompleteness:  75.0,
		Enabled:          true,
	}
	db.Create(mappingPM25)

	mappingPM10 := &model.ISPUParameterMapping{
		DeviceID:         dev.ID,
		ParameterID:      paramPM10.ID,
		PollutantCode:    model.PollutantPM10,
		SourceUnit:       "µg/m³",
		NormalizedUnit:   "µg/m³",
		ConversionFactor: 1.0,
		MinCompleteness:  75.0,
		Enabled:          true,
	}
	db.Create(mappingPM10)

	ispuRepo := repository.NewISPURepository(db)
	sysRepo := repository.NewSystemRepository(db)
	ispuSvc := service.NewISPUService(db, ispuRepo, sysRepo)

	return db, ispuSvc, dev
}

func TestISPUService_CalculateStationWindowAndPersistence(t *testing.T) {
	db, ispuSvc, dev := setupISPUTestDB(t)
	ctx := context.Background()

	now := time.Now().UTC()

	// Seed 20 hours of valid hourly telemetry (20 / 24 = 83.3% completeness -> VALID)
	// PM2.5 concentration = 31.4 µg/m³ -> ISPU 70 (Sedang)
	// PM10 concentration = 40.0 µg/m³ -> ISPU 40 (Baik)
	var pm25Param, pm10Param model.Parameter
	db.Where("device_id = ? AND parameter_code = 'PM25'", dev.ID).First(&pm25Param)
	db.Where("device_id = ? AND parameter_code = 'PM10'", dev.ID).First(&pm10Param)

	for h := 0; h < 20; h++ {
		ts := now.Add(-time.Duration(h) * time.Hour).Add(-5 * time.Minute)
		db.Create(&model.RawData{
			DeviceID:       dev.ID,
			ParameterID:    pm25Param.ID,
			ProcessedValue: 31.4,
			Quality:        model.QualityGood,
			ReceivedAt:     ts,
		})
		db.Create(&model.RawData{
			DeviceID:       dev.ID,
			ParameterID:    pm10Param.ID,
			ProcessedValue: 40.0,
			Quality:        model.QualityGood,
			ReceivedAt:     ts,
		})
	}

	result, err := ispuSvc.CalculateStationWindow(ctx, dev.ID, now)
	if err != nil {
		t.Fatalf("CalculateStationWindow failed: %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil ISPUCalculationResult")
	}

	// Overall ISPU should be max(PM2.5=70, PM10=40) = 70 (Sedang)
	if result.OverallCategory != model.CategorySedang {
		t.Errorf("Expected Category Sedang, got %s", result.OverallCategory)
	}
	if result.DominantPollutant != model.PollutantPM25 {
		t.Errorf("Expected Dominant Pollutant PM25, got %s", result.DominantPollutant)
	}
	if result.CalculationStatus != model.ISPUStatusValid {
		t.Errorf("Expected CalculationStatus VALID (83.3%% coverage), got %s", result.CalculationStatus)
	}

	// Verify persistence in DB
	var dbResult model.ISPUCalculationResult
	err = db.Preload("ParameterResults").First(&dbResult, result.ID).Error
	if err != nil {
		t.Fatalf("Failed to fetch persisted result from DB: %v", err)
	}
	if len(dbResult.ParameterResults) != 2 {
		t.Errorf("Expected 2 parameter results persisted, got %d", len(dbResult.ParameterResults))
	}

	// Test idempotency: re-running for the same period should update, not duplicate
	result2, err := ispuSvc.CalculateStationWindow(ctx, dev.ID, now)
	if err != nil {
		t.Fatalf("Second calculation run failed: %v", err)
	}
	if result2.ID != result.ID {
		t.Errorf("Expected idempotent update with same ID %d, got %d", result.ID, result2.ID)
	}
	var count int64
	db.Model(&model.ISPUCalculationResult{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected exactly 1 calculation result in DB after re-run, got %d", count)
	}
}

func TestISPUService_ExportCSVAndExcel(t *testing.T) {
	_, ispuSvc, dev := setupISPUTestDB(t)
	ctx := context.Background()

	// Calculate a record
	_, err := ispuSvc.CalculateStationWindow(ctx, dev.ID, time.Now())
	if err != nil {
		t.Fatalf("Calculation failed: %v", err)
	}

	// Export CSV
	csvData, err := ispuSvc.ExportCSV(ctx, repository.ISPUHistoryFilter{DeviceID: dev.ID})
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}
	csvStr := string(csvData)
	if !strings.Contains(csvStr, "AQMS-TEST-01") {
		t.Errorf("Expected station code in CSV export, got: %s", csvStr)
	}
	if !strings.Contains(csvStr, "PM2.5 ISPU") {
		t.Errorf("Expected pollutant headers in CSV export")
	}

	// Export Excel
	excelData, err := ispuSvc.ExportExcel(ctx, repository.ISPUHistoryFilter{DeviceID: dev.ID})
	if err != nil {
		t.Fatalf("ExportExcel failed: %v", err)
	}
	// Check UTF-8 BOM
	if len(excelData) < 3 || excelData[0] != 0xEF || excelData[1] != 0xBB || excelData[2] != 0xBF {
		t.Errorf("Expected UTF-8 BOM at start of Excel export")
	}
}

func TestISPUService_BackfillStationHourlyTrend(t *testing.T) {
	_, ispuSvc, dev := setupISPUTestDB(t)
	ctx := context.Background()

	// Backfill 24 hours of trend
	count, err := ispuSvc.BackfillStationHourlyTrend(ctx, dev.ID, 24)
	if err != nil {
		t.Fatalf("BackfillStationHourlyTrend failed: %v", err)
	}
	if count < 20 {
		t.Errorf("Expected at least 20 hourly windows backfilled, got %d", count)
	}

	// Re-run backfill to verify idempotency (no duplicate key errors)
	count2, err2 := ispuSvc.BackfillStationHourlyTrend(ctx, dev.ID, 24)
	if err2 != nil {
		t.Fatalf("Idempotent BackfillStationHourlyTrend failed: %v", err2)
	}
	if count2 != count {
		t.Errorf("Expected second run to process %d windows, got %d", count, count2)
	}

	// Verify trend query returns records
	trend, err := ispuSvc.GetTrend(ctx, dev.ID, "24h")
	if err != nil {
		t.Fatalf("GetTrend failed: %v", err)
	}
	if len(trend) < 20 {
		t.Errorf("Expected at least 20 trend records in 24h range, got %d", len(trend))
	}
}

func TestISPUService_CalculateFromAggregationResults(t *testing.T) {
	db, ispuSvc, dev := setupISPUTestDB(t)
	ctx := context.Background()

	now := time.Now().UTC()

	var pm25Param, pm10Param model.Parameter
	db.Where("device_id = ? AND parameter_code = 'PM25'", dev.ID).First(&pm25Param)
	db.Where("device_id = ? AND parameter_code = 'PM10'", dev.ID).First(&pm10Param)

	// Seed 20 hours of 2-minute aggregation results (10 buckets per hour for 20 hours)
	// PM2.5 avg = 56.4 µg/m³ -> 24h ISPU ~ 101 (Tidak Sehat)
	// PM10 avg = 40.0 µg/m³ -> 24h ISPU = 40 (Baik)
	valPM25 := 56.4
	valPM10 := 40.0
	for h := 0; h < 20; h++ {
		periodEnd := now.Add(-time.Duration(h) * time.Hour)
		periodStart := periodEnd.Add(-2 * time.Minute)
		db.Create(&model.AggregationResult{
			DeviceID:        dev.ID,
			ParameterID:     pm25Param.ID,
			SourceType:      model.SourceCustomerProcessed,
			PeriodStart:     periodStart,
			PeriodEnd:       periodEnd,
			IntervalSeconds: 120,
			AvgValue:        &valPM25,
			Quality:         model.QualityGood,
		})
		db.Create(&model.AggregationResult{
			DeviceID:        dev.ID,
			ParameterID:     pm10Param.ID,
			SourceType:      model.SourceCustomerProcessed,
			PeriodStart:     periodStart,
			PeriodEnd:       periodEnd,
			IntervalSeconds: 120,
			AvgValue:        &valPM10,
			Quality:         model.QualityGood,
		})
	}

	result, err := ispuSvc.CalculateStationWindow(ctx, dev.ID, now)
	if err != nil {
		t.Fatalf("CalculateStationWindow from aggregation failed: %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil ISPUCalculationResult")
	}

	// Overall 24h ISPU should be based on PM2.5 = 101 (Tidak Sehat)
	if result.OverallCategory != model.CategoryTidakSehat {
		t.Errorf("Expected Category Tidak Sehat, got %s", result.OverallCategory)
	}
	if result.DominantPollutant != model.PollutantPM25 {
		t.Errorf("Expected Dominant Pollutant PM25, got %s", result.DominantPollutant)
	}
	if result.OverallISPU < 100 || result.OverallISPU > 105 {
		t.Errorf("Expected OverallISPU ~101, got %.1f", result.OverallISPU)
	}

	// Verify instantaneous Aggregation ISPU
	if result.AggregationISPU < 100 || result.AggregationISPU > 105 {
		t.Errorf("Expected AggregationISPU ~101, got %.1f", result.AggregationISPU)
	}
	if result.AggregationCategory != model.CategoryTidakSehat {
		t.Errorf("Expected AggregationCategory Tidak Sehat, got %s", result.AggregationCategory)
	}
	if result.AggregationDominant != model.PollutantPM25 {
		t.Errorf("Expected AggregationDominant PM25, got %s", result.AggregationDominant)
	}

	// Verify parameter results have both 24h and aggregation values
	if len(result.ParameterResults) != 2 {
		t.Fatalf("Expected 2 parameter results, got %d", len(result.ParameterResults))
	}
	for _, pr := range result.ParameterResults {
		if pr.PollutantCode == model.PollutantPM25 {
			if pr.AggregationConcentration != 56.4 {
				t.Errorf("Expected PM25 AggregationConcentration 56.4, got %.1f", pr.AggregationConcentration)
			}
			if pr.AggregationISPU < 100 || pr.AggregationISPU > 105 {
				t.Errorf("Expected PM25 AggregationISPU ~101, got %.1f", pr.AggregationISPU)
			}
		}
	}
}

func TestISPUService_AlignToAggregationInterval(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	svc := &service.ISPUService{}

	// Test 2-minute interval (120s): all results MUST have even minutes (00, 02, 04 ... 36, 38, 40, 42)
	testCases2m := []struct {
		input    time.Time
		expected string
	}{
		{time.Date(2026, 10, 15, 15, 41, 36, 0, loc), "15:40:00"},
		{time.Date(2026, 10, 15, 15, 39, 59, 0, loc), "15:40:00"},
		{time.Date(2026, 10, 15, 15, 40, 0, 0, loc), "15:40:00"},
		{time.Date(2026, 10, 15, 15, 42, 0, 0, loc), "15:42:00"},
		{time.Date(2026, 10, 15, 15, 37, 12, 0, loc), "15:36:00"},
		{time.Date(2026, 10, 15, 15, 35, 0, 0, loc), "15:34:00"},
	}

	for _, tc := range testCases2m {
		aligned := svc.AlignToAggregationInterval(tc.input, 120, loc)
		actualStr := aligned.Format("15:04:05")
		if actualStr != tc.expected {
			t.Errorf("Input %s: expected aligned %s, got %s", tc.input.Format("15:04:05"), tc.expected, actualStr)
		}
		// Confirm minute is strictly even
		if aligned.Minute()%2 != 0 {
			t.Errorf("Aligned timestamp %s must be an even minute, got %d", actualStr, aligned.Minute())
		}
		if aligned.Second() != 0 {
			t.Errorf("Aligned timestamp %s must have 00 seconds, got %d", actualStr, aligned.Second())
		}
	}
}

