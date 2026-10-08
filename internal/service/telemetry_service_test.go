package service_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/internal/websocket"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// setupTelemetryTestDB creates an in-memory SQLite database using pure-Go glebarez/sqlite
func setupTelemetryTestDB(t *testing.T) (*gorm.DB, *repository.TelemetryRepository) {
	dsn := fmt.Sprintf("file:mem_telem_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.RawData{},
	)
	if err != nil {
		t.Fatalf("Failed to auto-migrate models: %v", err)
	}

	repo := repository.NewTelemetryRepository(db)
	return db, repo
}

// 1. Test Telemetry Model, Types & Sanitization
func TestTelemetryModelAndSanitization(t *testing.T) {
	db, repo := setupTelemetryTestDB(t)
	_ = db
	svc := service.NewTelemetryService(repo, nil, service.TelemetryConfig{
		BufferSize:    100,
		BatchSize:     10,
		FlushInterval: 50 * time.Millisecond,
	})
	defer svc.Stop()

	// 1a. Numeric Value
	payloadNum := &service.TelemetryIngestPayload{
		DeviceID:      1,
		DeviceCode:    "DEV-01",
		ParameterID:   1,
		ParameterCode: "TEMP_01",
		DataType:      model.DataTypeFloat32,
		Value:         28.456,
		RawValue:      2845,
		Quality:       model.QualityGood,
	}
	err := svc.Ingest(payloadNum)
	if err != nil {
		t.Fatalf("Failed to ingest numeric telemetry: %v", err)
	}

	latest, err := svc.GetLatestForParameter(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Failed to get latest telemetry: %v", err)
	}
	if latest.Value != 28.456 {
		t.Errorf("Expected value 28.456, got %f", latest.Value)
	}
	if latest.ValueText != "28.46" {
		t.Errorf("Expected value_text '28.46', got '%s'", latest.ValueText)
	}
	if latest.Quality != model.QualityGood {
		t.Errorf("Expected quality GOOD, got %s", latest.Quality)
	}
	if latest.ReceivedAt.IsZero() {
		t.Errorf("ReceivedAt should not be zero")
	}

	// 1b. Boolean Value (True / ON)
	payloadBoolOn := &service.TelemetryIngestPayload{
		DeviceID:      1,
		ParameterID:   2,
		ParameterCode: "PUMP_STATUS",
		DataType:      model.DataTypeBoolean,
		Value:         1.0,
		Quality:       model.QualityGood,
	}
	err = svc.Ingest(payloadBoolOn)
	if err != nil {
		t.Fatalf("Failed to ingest bool telemetry: %v", err)
	}
	latestBool, err := svc.GetLatestForParameter(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("Failed to get latest bool: %v", err)
	}
	if latestBool.ValueBool == nil || !*latestBool.ValueBool {
		t.Errorf("Expected ValueBool true")
	}
	if latestBool.ValueText != "ON" {
		t.Errorf("Expected ValueText ON, got %s", latestBool.ValueText)
	}

	// 1c. Boolean Value (False / OFF)
	payloadBoolOff := &service.TelemetryIngestPayload{
		DeviceID:      1,
		ParameterID:   3,
		ParameterCode: "VALVE_STATUS",
		DataType:      model.DataTypeBoolean,
		Value:         0.0,
		Quality:       model.QualityGood,
	}
	err = svc.Ingest(payloadBoolOff)
	if err != nil {
		t.Fatalf("Failed to ingest bool off: %v", err)
	}
	latestOff, err := svc.GetLatestForParameter(context.Background(), 1, 3)
	if err != nil {
		t.Fatalf("Failed to get latest bool off: %v", err)
	}
	if latestOff.ValueBool == nil || *latestOff.ValueBool {
		t.Errorf("Expected ValueBool false")
	}
	if latestOff.ValueText != "OFF" {
		t.Errorf("Expected ValueText OFF, got %s", latestOff.ValueText)
	}

	// 1d. Integer Rounding
	payloadInt := &service.TelemetryIngestPayload{
		DeviceID:      1,
		ParameterID:   4,
		ParameterCode: "RPM",
		DataType:      model.DataTypeUInt16,
		Value:         1450.7,
		Quality:       model.QualityGood,
	}
	_ = svc.Ingest(payloadInt)
	latestInt, err := svc.GetLatestForParameter(context.Background(), 1, 4)
	if err != nil {
		t.Fatalf("Failed to get latest int: %v", err)
	}
	if latestInt.ValueText != "1451" {
		t.Errorf("Expected ValueText '1451', got '%s'", latestInt.ValueText)
	}

	// 1e. NaN and Inf Sanitization
	payloadNaN := &service.TelemetryIngestPayload{
		DeviceID:      1,
		ParameterID:   5,
		ParameterCode: "FLOW",
		DataType:      model.DataTypeFloat32,
		Value:         math.NaN(),
		RawValue:      math.Inf(1),
		Quality:       model.QualityGood,
	}
	_ = svc.Ingest(payloadNaN)
	latestNaN, err := svc.GetLatestForParameter(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("Failed to get latest NaN: %v", err)
	}
	if latestNaN.Value != 0.0 {
		t.Errorf("Expected sanitized value 0.0, got %f", latestNaN.Value)
	}
	if latestNaN.Quality != model.QualityBad {
		t.Errorf("Expected Quality BAD on NaN, got %s", latestNaN.Quality)
	}
}

// 2. Test Ingestion Service & Instantaneous In-Memory Latest Cache
func TestTelemetryServiceIngestAndLatestCache(t *testing.T) {
	db, repo := setupTelemetryTestDB(t)
	svc := service.NewTelemetryService(repo, nil, service.TelemetryConfig{
		BufferSize:    1000,
		BatchSize:     50,
		FlushInterval: 100 * time.Millisecond,
	})
	defer svc.Stop()

	// Seed parameter in DB
	param := model.Parameter{
		ID:            10,
		DeviceID:      1,
		ParameterCode: "PRESSURE_01",
		ParameterName: "Hydraulic Pressure",
		Unit:          "bar",
		DataType:      model.DataTypeFloat32,
	}
	if err := db.Create(&param).Error; err != nil {
		t.Fatalf("Failed to seed parameter: %v", err)
	}

	// Ingest telemetry measurement
	err := svc.Ingest(&service.TelemetryIngestPayload{
		DeviceID:      1,
		DeviceCode:    "DEV-01",
		DeviceName:    "Main Hydraulic Pack",
		ParameterID:   10,
		ParameterCode: "PRESSURE_01",
		ParameterName: "Hydraulic Pressure",
		Unit:          "bar",
		DataType:      model.DataTypeFloat32,
		Value:         165.25,
		RawValue:      1652,
		RawHex:        "06 74",
		Quality:       model.QualityGood,
		Source:        "MODBUS_TCP",
		ReceivedAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Failed to ingest measurement: %v", err)
	}

	// Verify instantaneous O(1) in-memory cache lookup
	latest, err := svc.GetLatestForParameter(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("Failed to get latest parameter: %v", err)
	}
	if latest.Value != 165.25 {
		t.Errorf("Expected value 165.25, got %f", latest.Value)
	}
	if latest.Unit != "bar" {
		t.Errorf("Expected unit bar, got %s", latest.Unit)
	}
	if latest.RawHex != "06 74" {
		t.Errorf("Expected RawHex '06 74', got '%s'", latest.RawHex)
	}
	if latest.Quality != model.QualityGood {
		t.Errorf("Expected Quality GOOD, got %s", latest.Quality)
	}
	if latest.Source != "MODBUS_TCP" {
		t.Errorf("Expected Source MODBUS_TCP, got %s", latest.Source)
	}

	// Verify device list latest lookup
	allLatest, err := svc.GetLatestForDevice(context.Background(), 1)
	if err != nil {
		t.Fatalf("Failed to get latest for device: %v", err)
	}
	if len(allLatest) != 1 {
		t.Fatalf("Expected 1 latest item, got %d", len(allLatest))
	}
	if allLatest[0].Value != 165.25 {
		t.Errorf("Expected value 165.25, got %f", allLatest[0].Value)
	}
}

// 3. Test Realtime WebSocket Telemetry Broadcasting
func TestTelemetryWebSocketBroadcast(t *testing.T) {
	_, repo := setupTelemetryTestDB(t)

	hub := websocket.NewHub()
	go hub.Run()

	svc := service.NewTelemetryService(repo, hub, service.TelemetryConfig{
		BufferSize:    100,
		BatchSize:     10,
		FlushInterval: 100 * time.Millisecond,
	})
	defer svc.Stop()

	// Ingest telemetry with hub active
	err := svc.Ingest(&service.TelemetryIngestPayload{
		DeviceID:      2,
		DeviceCode:    "DEV-02",
		ParameterID:   20,
		ParameterCode: "VOLTAGE_L1",
		DataType:      model.DataTypeFloat32,
		Value:         230.5,
		Quality:       model.QualityGood,
		ReceivedAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Failed to ingest with hub: %v", err)
	}

	metrics := svc.GetMetrics()
	if metrics.IngestedCount != 1 {
		t.Errorf("Expected IngestedCount 1, got %d", metrics.IngestedCount)
	}
}

// 4. Test Buffered Batch Persistence & Graceful Shutdown Zero Data Loss
func TestTelemetryBatchPersistenceAndShutdownFlush(t *testing.T) {
	db, repo := setupTelemetryTestDB(t)

	dev := model.Device{ID: 1, DeviceCode: "DEV-BATCH", DeviceName: "Batch Device"}
	if err := db.Create(&dev).Error; err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}
	param := model.Parameter{ID: 1, DeviceID: 1, ParameterCode: "VAL_01"}
	if err := db.Create(&param).Error; err != nil {
		t.Fatalf("Failed to create parameter: %v", err)
	}

	svc := service.NewTelemetryService(repo, nil, service.TelemetryConfig{
		BufferSize:    1000,
		BatchSize:     50,
		FlushInterval: 500 * time.Millisecond, // long interval so we test shutdown flush
	})

	// Ingest 25 records (less than batch size of 50)
	for i := 1; i <= 25; i++ {
		err := svc.Ingest(&service.TelemetryIngestPayload{
			DeviceID:      1,
			ParameterID:   1,
			ParameterCode: "VAL_01",
			Value:         float64(i * 10),
			Quality:       model.QualityGood,
			ReceivedAt:    time.Now().UTC().Add(time.Duration(i) * time.Millisecond),
		})
		if err != nil {
			t.Fatalf("Failed to ingest batch item %d: %v", i, err)
		}
	}

	// Trigger graceful stop (must drain and flush all 25 records to SQLite)
	svc.Stop()

	// Verify all 25 records were committed to database
	var count int64
	if err := db.Model(&model.RawData{}).Where("device_id = ?", 1).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count raw data: %v", err)
	}
	if count != 25 {
		t.Fatalf("Expected 25 persisted records, got %d", count)
	}

	// Verify parameter latest value updated
	var updatedParam model.Parameter
	if err := db.First(&updatedParam, 1).Error; err != nil {
		t.Fatalf("Failed to fetch updated parameter: %v", err)
	}
	if updatedParam.CurrentValue == nil || *updatedParam.CurrentValue != 250.0 {
		t.Errorf("Expected latest parameter value 250.0, got %v", updatedParam.CurrentValue)
	}
}

// 5. Test Bounded Buffer and Non-Blocking Backpressure Protection
func TestTelemetryBoundedBufferBackpressure(t *testing.T) {
	_, repo := setupTelemetryTestDB(t)

	// Tiny buffer capacity of 5 items with very long flush interval
	svc := service.NewTelemetryService(repo, nil, service.TelemetryConfig{
		BufferSize:    5,
		BatchSize:     100,
		FlushInterval: 10 * time.Second,
	})
	defer svc.Stop()

	// Rapidly ingest 50 items to overflow the tiny buffer of 5
	dropped := 0
	for i := 1; i <= 50; i++ {
		err := svc.Ingest(&service.TelemetryIngestPayload{
			DeviceID:      1,
			ParameterID:   uint(i),
			ParameterCode: "PARAM",
			Value:         float64(i),
			Quality:       model.QualityGood,
		})
		if err != nil {
			dropped++
		}
	}

	metrics := svc.GetMetrics()
	if metrics.IngestedCount != 50 {
		t.Errorf("Expected IngestedCount 50, got %d", metrics.IngestedCount)
	}
	if metrics.DroppedCount == 0 {
		t.Errorf("Expected dropped count > 0 under backpressure")
	}
	if metrics.DroppedCount != uint64(dropped) {
		t.Errorf("Dropped count mismatch: %d vs %d", metrics.DroppedCount, dropped)
	}
}

// 6. Test Multi-Device Ingestion Isolation
func TestMultiDeviceIngestionIsolation(t *testing.T) {
	db, repo := setupTelemetryTestDB(t)

	svc := service.NewTelemetryService(repo, nil, service.TelemetryConfig{
		BufferSize:    5000,
		BatchSize:     20,
		FlushInterval: 50 * time.Millisecond,
	})
	defer svc.Stop()

	// Register 5 devices and parameters
	for d := uint(1); d <= 5; d++ {
		if err := db.Create(&model.Device{ID: d, DeviceCode: fmt.Sprintf("DEV-%d", d)}).Error; err != nil {
			t.Fatalf("Failed to create device %d: %v", d, err)
		}
		if err := db.Create(&model.Parameter{ID: d, DeviceID: d, ParameterCode: "VAL"}).Error; err != nil {
			t.Fatalf("Failed to create parameter %d: %v", d, err)
		}
	}

	// Concurrently ingest from 5 distinct device channels
	var wg sync.WaitGroup
	for d := uint(1); d <= 5; d++ {
		wg.Add(1)
		go func(deviceID uint) {
			defer wg.Done()
			for i := 1; i <= 20; i++ {
				_ = svc.Ingest(&service.TelemetryIngestPayload{
					DeviceID:      deviceID,
					ParameterID:   deviceID,
					ParameterCode: "VAL",
					Value:         float64(deviceID*1000 + uint(i)),
					Quality:       model.QualityGood,
					ReceivedAt:    time.Now().UTC(),
				})
			}
		}(d)
	}
	wg.Wait()

	// Allow flush
	time.Sleep(150 * time.Millisecond)

	// Verify each device has isolated records
	for d := uint(1); d <= 5; d++ {
		var count int64
		if err := db.Model(&model.RawData{}).Where("device_id = ?", d).Count(&count).Error; err != nil {
			t.Fatalf("Failed to count device %d raw data: %v", d, err)
		}
		if count != 20 {
			t.Errorf("Expected 20 records for device %d, got %d", d, count)
		}
	}
}

// 7. Mock Failing Repository to Test Transient DB Retry
type MockFailingRepo struct {
	attempts int32
	failMax  int32
}

func (m *MockFailingRepo) SaveBatch(ctx context.Context, batch []*model.RawData) error {
	attempt := atomic.AddInt32(&m.attempts, 1)
	if attempt <= m.failMax {
		return errors.New("simulated transient MariaDB deadlock/timeout")
	}
	return nil
}

func TestTelemetryDatabaseRetryLogic(t *testing.T) {
	mockRepo := &MockFailingRepo{failMax: 2} // fails first 2 attempts, succeeds on 3rd

	batch := []*model.RawData{
		{DeviceID: 1, ParameterID: 1, Value: 42.0},
	}

	start := time.Now()
	var err error
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = mockRepo.SaveBatch(context.Background(), batch)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("Expected retry to eventually succeed, got: %v", err)
	}
	if atomic.LoadInt32(&mockRepo.attempts) != 3 {
		t.Errorf("Expected 3 attempts, got %d", mockRepo.attempts)
	}
	if time.Since(start) < 20*time.Millisecond {
		t.Errorf("Expected backoff delay")
	}
}

// 8. Test Historical Telemetry Filtering, Pagination & Retention
func TestTelemetryHistoricalFilteringAndPagination(t *testing.T) {
	db, repo := setupTelemetryTestDB(t)

	baseTime := time.Date(2026, 10, 7, 10, 0, 0, 0, time.UTC)

	// Seed 30 records:
	// - 15 records for Device 1, Parameter 1, GOOD quality
	// - 10 records for Device 1, Parameter 2, BAD quality
	// - 5 records for Device 2, Parameter 3, UNCERTAIN quality
	var seedData []*model.RawData
	for i := 1; i <= 15; i++ {
		seedData = append(seedData, &model.RawData{
			DeviceID:    1,
			ParameterID: 1,
			Value:       float64(i),
			Quality:     model.QualityGood,
			Source:      "MODBUS_TCP",
			ReceivedAt:  baseTime.Add(time.Duration(i) * time.Minute),
			StoredAt:    baseTime.Add(time.Duration(i) * time.Minute),
			Timestamp:   baseTime.Add(time.Duration(i) * time.Minute),
		})
	}
	for i := 1; i <= 10; i++ {
		seedData = append(seedData, &model.RawData{
			DeviceID:    1,
			ParameterID: 2,
			Value:       float64(i * 10),
			Quality:     model.QualityBad,
			Source:      "MODBUS_TCP",
			ReceivedAt:  baseTime.Add(time.Duration(i+15) * time.Minute),
			StoredAt:    baseTime.Add(time.Duration(i+15) * time.Minute),
			Timestamp:   baseTime.Add(time.Duration(i+15) * time.Minute),
		})
	}
	for i := 1; i <= 5; i++ {
		seedData = append(seedData, &model.RawData{
			DeviceID:    2,
			ParameterID: 3,
			Value:       float64(i * 100),
			Quality:     model.QualityUncertain,
			Source:      "MODBUS_RTU",
			ReceivedAt:  baseTime.Add(time.Duration(i+25) * time.Minute),
			StoredAt:    baseTime.Add(time.Duration(i+25) * time.Minute),
			Timestamp:   baseTime.Add(time.Duration(i+25) * time.Minute),
		})
	}
	if err := db.Create(&seedData).Error; err != nil {
		t.Fatalf("Failed to seed raw data: %v", err)
	}

	// 8a. Filter by Parameter ID
	records, total, err := repo.GetHistorical(context.Background(), repository.TelemetryFilterParams{
		DeviceID:    1,
		ParameterID: 1,
		Page:        1,
		PageSize:    50,
	})
	if err != nil {
		t.Fatalf("Failed to get historical for param 1: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected 15 records, got %d", total)
	}
	if len(records) != 15 {
		t.Errorf("Expected 15 records slice, got %d", len(records))
	}

	// 8b. Filter by Quality
	badRecords, badTotal, err := repo.GetHistorical(context.Background(), repository.TelemetryFilterParams{
		DeviceID: 1,
		Quality:  "BAD",
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("Failed to get BAD quality records: %v", err)
	}
	if badTotal != 10 {
		t.Errorf("Expected 10 BAD records, got %d", badTotal)
	}
	if len(badRecords) != 10 {
		t.Errorf("Expected 10 BAD records slice, got %d", len(badRecords))
	}

	// 8c. Pagination Check (Page 1 with PageSize 5 out of 15)
	page1, total1, err := repo.GetHistorical(context.Background(), repository.TelemetryFilterParams{
		DeviceID:    1,
		ParameterID: 1,
		Page:        1,
		PageSize:    5,
	})
	if err != nil {
		t.Fatalf("Failed page 1: %v", err)
	}
	if total1 != 15 {
		t.Errorf("Expected total 15, got %d", total1)
	}
	if len(page1) != 5 {
		t.Errorf("Expected 5 items on page 1, got %d", len(page1))
	}

	// Page 2
	page2, _, err := repo.GetHistorical(context.Background(), repository.TelemetryFilterParams{
		DeviceID:    1,
		ParameterID: 1,
		Page:        2,
		PageSize:    5,
	})
	if err != nil {
		t.Fatalf("Failed page 2: %v", err)
	}
	if len(page2) != 5 {
		t.Errorf("Expected 5 items on page 2, got %d", len(page2))
	}
	if page1[0].ID == page2[0].ID {
		t.Errorf("Page 1 and Page 2 should have distinct records")
	}

	// 8d. Time Range Filter (between Minute 5 and Minute 10)
	t1 := baseTime.Add(5 * time.Minute)
	t2 := baseTime.Add(10 * time.Minute)
	rangeRecords, rangeTotal, err := repo.GetHistorical(context.Background(), repository.TelemetryFilterParams{
		DeviceID:    1,
		ParameterID: 1,
		StartTime:   &t1,
		EndTime:     &t2,
		Page:        1,
		PageSize:    50,
	})
	if err != nil {
		t.Fatalf("Failed time range query: %v", err)
	}
	if rangeTotal != 6 {
		t.Errorf("Expected 6 items in range, got %d", rangeTotal)
	}
	if len(rangeRecords) != 6 {
		t.Errorf("Expected 6 items slice, got %d", len(rangeRecords))
	}

	// 8e. Data Retention Foundation: Delete older than cutoff
	cutoff := baseTime.Add(12 * time.Minute)
	deleted, err := repo.DeleteOlderThan(context.Background(), cutoff)
	if err != nil {
		t.Fatalf("Failed to delete older than: %v", err)
	}
	if deleted <= 0 {
		t.Errorf("Expected deleted > 0, got %d", deleted)
	}

	var remaining int64
	if err := db.Model(&model.RawData{}).Count(&remaining).Error; err != nil {
		t.Fatalf("Failed to count remaining: %v", err)
	}
	if remaining != int64(30)-deleted {
		t.Errorf("Expected %d remaining, got %d", int64(30)-deleted, remaining)
	}
}
