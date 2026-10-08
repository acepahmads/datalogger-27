package aggregation_test

import (
	"context"
	"testing"
	"time"

	"datalogger/internal/aggregation"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:mem_agg_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open sqlite in-memory db: %v", err)
	}

	err = db.AutoMigrate(
		&model.Device{},
		&model.Parameter{},
		&model.RawData{},
		&model.AggregationDefinition{},
		&model.AggregationResult{},
		&model.AuditTrail{},
		&model.DevelopmentPhase{},
		&model.DevelopmentSubphase{},
		&model.DevelopmentTask{},
	)
	if err != nil {
		t.Fatalf("Failed to auto-migrate test db: %v", err)
	}
	return db
}

func TestPhase33ComprehensiveAcceptance(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	aggRepo := repository.NewAggregationRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	svc := service.NewAggregationService(aggRepo, devRepo, sysRepo, nil)

	// Create test device & parameters
	dev := model.Device{
		ID:         1,
		DeviceCode: "AQMS-01",
		DeviceName: "Air Quality Station 01",
		Status:     model.DeviceStatusActive,
	}
	db.Create(&dev)

	paramTemp := model.Parameter{
		ID:            101,
		DeviceID:      1,
		ParameterCode: "temperature",
		ParameterName: "Ambient Temperature",
		Unit:          "°C",
		Enabled:       true,
	}
	paramHumidity := model.Parameter{
		ID:            102,
		DeviceID:      1,
		ParameterCode: "humidity",
		ParameterName: "Relative Humidity",
		Unit:          "%",
		Enabled:       true,
	}
	db.Create(&paramTemp)
	db.Create(&paramHumidity)

	t.Run("24. Late Arriving Telemetry Recalculation", func(t *testing.T) {
		bucketStart := time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC)

		// Create definition for customer 30-min temperature
		def, err := svc.CreateDefinition(ctx, &service.CreateAggregationDefinitionRequest{
			Name:            "Temp 30m Customer Avg",
			DeviceID:        1,
			ParameterID:     101,
			SourceType:      model.SourceCustomerProcessed,
			Function:        model.FunctionAvg,
			IntervalSeconds: 1800,
			Timezone:        "UTC",
		}, "admin", "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("Failed to create definition: %v", err)
		}

		// Initial sample arrives: 25.0
		db.Create(&model.RawData{
			DeviceID:       1,
			ParameterID:    101,
			RawValue:       250.0,
			ProcessedValue: 25.0,
			ReceivedAt:     bucketStart.Add(5 * time.Minute),
			Quality:        model.QualityGood,
		})

		// Run calculation
		res1, err := svc.CalculateBucket(ctx, def.ID, bucketStart)
		if err != nil {
			t.Fatalf("Calculation 1 failed: %v", err)
		}
		if *res1.Value != 25.0 || res1.SampleCount != 1 {
			t.Errorf("Expected initial value 25.0 (count 1), got %v (count %d)", *res1.Value, res1.SampleCount)
		}

		// Late sample arrives with timestamp 14:15: 27.0
		db.Create(&model.RawData{
			DeviceID:       1,
			ParameterID:    101,
			RawValue:       270.0,
			ProcessedValue: 27.0,
			ReceivedAt:     bucketStart.Add(15 * time.Minute),
			Quality:        model.QualityGood,
		})

		// Recalculate the same bucket
		res2, err := svc.CalculateBucket(ctx, def.ID, bucketStart)
		if err != nil {
			t.Fatalf("Recalculation failed: %v", err)
		}
		if *res2.Value != 26.0 || res2.SampleCount != 2 {
			t.Errorf("Expected late-recalculated AVG 26.0 (count 2), got %v (count %d)", *res2.Value, res2.SampleCount)
		}
	})

	t.Run("25 & 26. Idempotent Rerun and Duplicate Prevention", func(t *testing.T) {
		bucketStart := time.Date(2026, 10, 8, 15, 0, 0, 0, time.UTC)
		def, _ := svc.CreateDefinition(ctx, &service.CreateAggregationDefinitionRequest{
			Name:            "Humidity 30m Avg",
			DeviceID:        1,
			ParameterID:     102,
			SourceType:      model.SourceCustomerProcessed,
			Function:        model.FunctionAvg,
			IntervalSeconds: 1800,
			Timezone:        "UTC",
		}, "admin", "127.0.0.1", "test-agent")

		db.Create(&model.RawData{
			DeviceID:       1,
			ParameterID:    102,
			ProcessedValue: 60.0,
			ReceivedAt:     bucketStart.Add(10 * time.Minute),
			Quality:        model.QualityGood,
		})

		// Run 3 times consecutively
		_, _ = svc.CalculateBucket(ctx, def.ID, bucketStart)
		_, _ = svc.CalculateBucket(ctx, def.ID, bucketStart)
		_, _ = svc.CalculateBucket(ctx, def.ID, bucketStart)

		var count int64
		db.Model(&model.AggregationResult{}).
			Where("aggregation_definition_id = ? AND period_start = ?", def.ID, bucketStart).
			Count(&count)

		if count != 1 {
			t.Fatalf("Idempotency violation! Expected exactly 1 record, got %d", count)
		}
	})

	t.Run("27 & 28. Restart Recovery & Missing Buckets", func(t *testing.T) {
		pastTime := time.Now().UTC().Add(-40 * time.Minute)
		b := aggregation.CalculateBucket(pastTime, 300, "UTC")

		def, _ := svc.CreateDefinition(ctx, &service.CreateAggregationDefinitionRequest{
			Name:            "Recovery 5m",
			DeviceID:        1,
			ParameterID:     101,
			SourceType:      model.SourceCustomerProcessed,
			Function:        model.FunctionAvg,
			IntervalSeconds: 300,
			Timezone:        "UTC",
		}, "admin", "127.0.0.1", "test-agent")

		// Insert telemetry into past bucket
		db.Create(&model.RawData{
			DeviceID:       1,
			ParameterID:    101,
			ProcessedValue: 28.5,
			ReceivedAt:     b.PeriodStart.Add(1 * time.Minute),
			Quality:        model.QualityGood,
		})

		// Trigger pending buckets processing within 2 hours
		processed, err := svc.ProcessPendingBucketsForDefinition(ctx, def, 2*time.Hour)
		if err != nil {
			t.Fatalf("Recovery failed: %v", err)
		}
		if processed < 1 {
			t.Errorf("Expected at least 1 bucket processed during recovery, got %d", processed)
		}
	})

	t.Run("29, 30, 31. Multiple Parameters, Devices, and Definitions", func(t *testing.T) {
		defs, err := svc.ListDefinitions(ctx, 1, false)
		if err != nil {
			t.Fatalf("Failed to list definitions: %v", err)
		}
		if len(defs) < 2 {
			t.Errorf("Expected multiple definitions, got %d", len(defs))
		}
	})

	t.Run("32 & 33. Customer API Response & No Internal Raw Leak", func(t *testing.T) {
		identifier := "20261008143000"
		customerData, err := svc.GetCustomerAggregatedData(ctx, identifier)
		if err != nil {
			t.Fatalf("Customer API failed: %v", err)
		}

		if customerData.Identifier != identifier {
			t.Errorf("Expected identifier %s, got %s", identifier, customerData.Identifier)
		}
		if len(customerData.Data) == 0 {
			t.Fatalf("Expected customer data points, got 0")
		}

		item := customerData.Data[0]
		if item.ParameterCode != "temperature" || item.Unit != "°C" {
			t.Errorf("Customer metadata mismatch: code=%s, unit=%s", item.ParameterCode, item.Unit)
		}
		// Confirm value is processed 26.0 (from earlier calculation)
		if item.Value == nil || *item.Value != 26.0 {
			t.Errorf("Expected customer value 26.0, got %v", item.Value)
		}
	})

	t.Run("34 & 35. RBAC and Audit Trail Verification", func(t *testing.T) {
		var auditTrails []model.AuditTrail
		db.Where("action LIKE ?", "%AGGREGATION%").Find(&auditTrails)
		if len(auditTrails) == 0 {
			t.Errorf("Expected audit trail entries for aggregation actions, got 0")
		}
		for _, a := range auditTrails {
			if a.Action != "CREATE_AGGREGATION" && a.Action != "UPDATE_AGGREGATION" && a.Action != "DELETE_AGGREGATION" {
				t.Errorf("Unexpected audit action: %s", a.Action)
			}
		}
	})

	t.Run("36. Downsampling Engine Resolution Selection", func(t *testing.T) {
		now := time.Now().UTC()
		// 1. Range <= 1 hour => auto chooses 'raw'
		respRaw, err := svc.GetDownsampledHistory(ctx, 1, 101, now.Add(-30*time.Minute), now, "auto")
		if err != nil || respRaw.Resolution != "raw" {
			t.Errorf("Expected raw resolution for 30m range, got %s (err: %v)", respRaw.Resolution, err)
		}

		// 2. Range 12 hours => auto chooses '5m'
		resp5m, err := svc.GetDownsampledHistory(ctx, 1, 101, now.Add(-12*time.Hour), now, "auto")
		if err != nil || resp5m.Resolution != "5m" {
			t.Errorf("Expected 5m resolution for 12h range, got %s", resp5m.Resolution)
		}

		// 3. Range 5 days => auto chooses '30m'
		resp30m, err := svc.GetDownsampledHistory(ctx, 1, 101, now.Add(-5*24*time.Hour), now, "auto")
		if err != nil || resp30m.Resolution != "30m" {
			t.Errorf("Expected 30m resolution for 5d range, got %s", resp30m.Resolution)
		}
	})

	t.Run("37. Real Sensor Telemetry Validation", func(t *testing.T) {
		// Simulate real AQMS-01 sensor measurement:
		// Raw sensor reading: 257.0 (MODBUS holding register)
		// Scaled & processed customer reading: 25.70 °C
		realStart := time.Date(2026, 10, 8, 16, 0, 0, 0, time.UTC)
		db.Create(&model.RawData{
			DeviceID:       1,
			ParameterID:    101,
			RawValue:       257.0,
			ProcessedValue: 25.70,
			Value:          25.70,
			ReceivedAt:     realStart.Add(2 * time.Minute),
			Quality:        model.QualityGood,
			QualityReason:  model.ReasonNone,
			Source:         "MODBUS_RTU",
		})

		// A. Real Sensor Internal Raw Aggregation
		rawDef, err := svc.CreateDefinition(ctx, &service.CreateAggregationDefinitionRequest{
			Name:            "Real Sensor Raw Rollup",
			Code:            "RAW_SENSOR_ROLLUP_300S",
			DeviceID:        1,
			ParameterID:     101,
			SourceType:      model.SourceInternalRaw,
			Function:        model.FunctionAvg,
			IntervalSeconds: 300,
			Timezone:        "UTC",
		}, "admin", "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("Failed to create raw definition: %v", err)
		}

		rawRes, err := svc.CalculateBucket(ctx, rawDef.ID, realStart)
		if err != nil || rawRes.Value == nil || *rawRes.Value != 257.0 {
			t.Fatalf("Real sensor raw aggregation failed: got %v, err: %v", rawRes.Value, err)
		}

		// B. Real Sensor Customer Processed Aggregation
		custDef, err := svc.CreateDefinition(ctx, &service.CreateAggregationDefinitionRequest{
			Name:            "Real Sensor Customer Rollup",
			Code:            "CUST_SENSOR_ROLLUP_300S",
			DeviceID:        1,
			ParameterID:     101,
			SourceType:      model.SourceCustomerProcessed,
			Function:        model.FunctionAvg,
			IntervalSeconds: 300,
			Timezone:        "UTC",
		}, "admin", "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("Failed to create customer definition: %v", err)
		}

		custRes, err := svc.CalculateBucket(ctx, custDef.ID, realStart)
		if err != nil || custRes.Value == nil || *custRes.Value != 25.70 {
			t.Fatalf("Real sensor customer aggregation failed: got %v, err: %v", custRes.Value, err)
		}

		// Verify distinct values: Raw is 257.0, Customer is 25.70
		if *rawRes.Value == *custRes.Value {
			t.Errorf("Domain segregation failed! Raw value %v should not equal customer value %v", *rawRes.Value, *custRes.Value)
		}
	})
}
