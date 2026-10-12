package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/handler"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupPackageATestEnv(t *testing.T) (*gin.Engine, *gorm.DB, *service.SystemService) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_pkg_a_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}

	_ = db.AutoMigrate(
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.Sensor{},
		&model.RawData{},
		&model.ProcessedData{},
		&model.Alarm{},
		&model.DevelopmentPhase{},
		&model.DevelopmentTask{},
	)

	cfg := &config.Config{
		Version: "v1.0.0",
		DBHost:  "127.0.0.1",
		DBPort:  "3306",
		DBName:  "datalogger",
	}

	systemRepo := repository.NewSystemRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)

	systemHandler := handler.NewSystemHandler(systemService)

	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/system/status", systemHandler.GetStatus)
		api.GET("/devices", systemHandler.GetDevices)
		api.GET("/devices/:id", systemHandler.GetDeviceByID)
	}

	return r, db, systemService
}

// TestPackageA_SystemStatusOverview verifies Task #67 Dashboard operational telemetry fields
func TestPackageA_SystemStatusOverview(t *testing.T) {
	r, db, _ := setupPackageATestEnv(t)

	// Seed test devices with known operational metrics
	now := time.Now()
	dev1 := model.Device{
		DeviceCode:        "AQMS-01",
		DeviceName:        "Air Quality Station 1",
		Status:            model.DeviceOnline,
		ConnectionStatus:  "ONLINE",
		LatencyMs:         14,
		SuccessCount:      1200,
		FailedCount:       2,
		LastCommunication: &now,
	}
	dev2 := model.Device{
		DeviceCode:        "AQMS-02",
		DeviceName:        "Air Quality Station 2",
		Status:            model.DeviceOffline,
		ConnectionStatus:  "OFFLINE",
		LatencyMs:         0,
		SuccessCount:      850,
		FailedCount:       45,
		LastCommunication: &now,
	}
	db.Create(&dev1)
	db.Create(&dev2)

	// Seed alarms (1 critical, 1 warning)
	db.Create(&model.Alarm{
		DeviceID:  &dev2.ID,
		AlarmCode: "DEVICE_OFFLINE",
		Severity:  model.SeverityCritical,
		Status:    model.AlarmActive,
		Message:   "Device AQMS-02 is OFFLINE",
	})
	db.Create(&model.Alarm{
		DeviceID:  &dev1.ID,
		AlarmCode: "HIGH_PM25",
		Severity:  model.SeverityWarning,
		Status:    model.AlarmActive,
		Message:   "PM2.5 above advisory limit",
	})

	req, _ := http.NewRequest("GET", "/api/system/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Success bool                       `json:"success"`
		Data    model.SystemStatusOverview `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if !res.Success {
		t.Fatalf("Expected success: true")
	}

	status := res.Data

	// 1. Verify Task #67 Dashboard operational data binding
	if status.TotalDevices != 2 {
		t.Errorf("Expected TotalDevices = 2, got %d", status.TotalDevices)
	}
	if status.OnlineDevices != 1 {
		t.Errorf("Expected OnlineDevices = 1, got %d", status.OnlineDevices)
	}
	if status.OfflineDevices != 1 {
		t.Errorf("Expected OfflineDevices = 1, got %d", status.OfflineDevices)
	}

	// 2. Verify latency is derived from real device data
	if status.AvgLatencyMs != 7 { // (14 + 0) / 2
		t.Errorf("Expected AvgLatencyMs = 7, got %d", status.AvgLatencyMs)
	}

	// 3. Verify real alarm metrics
	if status.ActiveAlarms != 2 {
		t.Errorf("Expected ActiveAlarms = 2, got %d", status.ActiveAlarms)
	}
	if status.CriticalAlarms != 1 {
		t.Errorf("Expected CriticalAlarms = 1, got %d", status.CriticalAlarms)
	}
	if status.WarningAlarms != 1 {
		t.Errorf("Expected WarningAlarms = 1, got %d", status.WarningAlarms)
	}

	// 4. Verify MariaDB database parameters
	if status.DBHost != "127.0.0.1" {
		t.Errorf("Expected DBHost = '127.0.0.1', got '%s'", status.DBHost)
	}
	if status.DBPort != "3306" {
		t.Errorf("Expected DBPort = '3306', got '%s'", status.DBPort)
	}
	if status.DBDatabase != "datalogger" {
		t.Errorf("Expected DBDatabase = 'datalogger', got '%s'", status.DBDatabase)
	}

	// 5. Verify dynamic throughput & quality calculation
	if status.DataReceivedCount < 2050 { // 1200 + 850
		t.Errorf("Expected DataReceivedCount >= 2050, got %d", status.DataReceivedCount)
	}
	if status.DataQuality == "" {
		t.Errorf("Expected non-empty DataQuality string")
	}
}

// TestPackageA_DeviceMonitoring verifies Task #68 Device Monitoring endpoint
func TestPackageA_DeviceMonitoring(t *testing.T) {
	r, db, _ := setupPackageATestEnv(t)

	now := time.Now()
	dev := model.Device{
		DeviceCode:        "MODBUS-DEV-01",
		DeviceName:        "Modbus Test Device",
		Status:            model.DeviceOnline,
		ConnectionStatus:  "ONLINE",
		LatencyMs:         8,
		SuccessCount:      500,
		FailedCount:       0,
		LastCommunication: &now,
	}
	db.Create(&dev)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/devices/%d", dev.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var res struct {
		Success bool         `json:"success"`
		Data    model.Device `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	if res.Data.DeviceCode != "MODBUS-DEV-01" {
		t.Errorf("Expected device code MODBUS-DEV-01, got %s", res.Data.DeviceCode)
	}
	if res.Data.ConnectionStatus != "ONLINE" {
		t.Errorf("Expected connection status ONLINE, got %s", res.Data.ConnectionStatus)
	}
	if res.Data.LatencyMs != 8 {
		t.Errorf("Expected latency 8ms, got %d", res.Data.LatencyMs)
	}
}

// TestPackageA_ParameterMonitoring verifies Task #69 Parameter Monitoring logic and quality indicators
func TestPackageA_ParameterMonitoring(t *testing.T) {
	r, db, _ := setupPackageATestEnv(t)

	now := time.Now()
	dev := model.Device{
		DeviceCode:        "AQMS-PARAM-TEST",
		DeviceName:        "Air Station Param Test",
		Status:            model.DeviceOnline,
		ConnectionStatus:  "ONLINE",
		LastCommunication: &now,
	}
	db.Create(&dev)

	val1 := 42.5
	val2 := 120.8
	formulaStr := "x * 1.05 + 2"
	param1 := model.Parameter{
		DeviceID:            dev.ID,
		ParameterCode:       "PM25",
		ParameterName:       "Particulate Matter 2.5",
		Unit:                "µg/m³",
		DataType:            model.DataTypeFloat32,
		CurrentValueNumeric: &val1,
		Formula:             formulaStr,
		CurrentFormulaValue: &val2,
		CurrentQuality:      model.QualityGood,
		Enabled:             true,
	}
	param2 := model.Parameter{
		DeviceID:       dev.ID,
		ParameterCode:  "TEMP",
		ParameterName:  "Ambient Temperature",
		Unit:           "°C",
		DataType:       model.DataTypeFloat32,
		CurrentQuality: model.QualityStale,
		Enabled:        true,
	}
	db.Create(&param1)
	db.Create(&param2)

	// Fetch device with preloaded parameters
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/devices/%d", dev.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var res struct {
		Success bool         `json:"success"`
		Data    model.Device `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	if len(res.Data.Parameters) != 2 {
		t.Fatalf("Expected 2 parameters preloaded, got %d", len(res.Data.Parameters))
	}

	p1 := res.Data.Parameters[0]
	if p1.ParameterCode != "PM25" {
		t.Errorf("Expected parameter code PM25, got %s", p1.ParameterCode)
	}
	if p1.Unit != "µg/m³" {
		t.Errorf("Expected unit µg/m³, got %s", p1.Unit)
	}
	if p1.CurrentQuality != model.QualityGood {
		t.Errorf("Expected quality GOOD, got %s", p1.CurrentQuality)
	}
	if p1.Formula != formulaStr {
		t.Errorf("Expected formula %s, got %s", formulaStr, p1.Formula)
	}

	p2 := res.Data.Parameters[1]
	if p2.CurrentQuality != model.QualityStale {
		t.Errorf("Expected quality STALE, got %s", p2.CurrentQuality)
	}
}

// TestPackageA_CommunicationMonitoring verifies Task #70 Communication Health aggregation
func TestPackageA_CommunicationMonitoring(t *testing.T) {
	r, db, _ := setupPackageATestEnv(t)

	now := time.Now()
	// Seed 3 devices with various communication stats
	d1 := model.Device{
		DeviceCode:        "COMM-01",
		DeviceName:        "Comm Link 1",
		Status:            model.DeviceOnline,
		ConnectionStatus:  "ONLINE",
		LatencyMs:         10,
		SuccessCount:      900,
		FailedCount:       100,
		LastCommunication: &now,
	}
	d2 := model.Device{
		DeviceCode:        "COMM-02",
		DeviceName:        "Comm Link 2",
		Status:            model.DeviceOnline,
		ConnectionStatus:  "ONLINE",
		LatencyMs:         20,
		SuccessCount:      950,
		FailedCount:       50,
		LastCommunication: &now,
	}
	d3 := model.Device{
		DeviceCode:        "COMM-03",
		DeviceName:        "Comm Link 3",
		Status:            model.DeviceOffline,
		ConnectionStatus:  "OFFLINE",
		LatencyMs:         0,
		SuccessCount:      0,
		FailedCount:       50,
		LastCommunication: &now,
	}
	db.Create(&d1)
	db.Create(&d2)
	db.Create(&d3)

	req, _ := http.NewRequest("GET", "/api/system/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var res struct {
		Success bool                       `json:"success"`
		Data    model.SystemStatusOverview `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	// Total connections = 3
	if res.Data.TotalConnections != 3 {
		t.Errorf("Expected TotalConnections = 3, got %d", res.Data.TotalConnections)
	}

	// Total comm = (900+100) + (950+50) + (0+50) = 1000 + 1000 + 50 = 2050
	// Total success = 900 + 950 = 1850
	// Success rate = 1850 / 2050 * 100 = 90.24%
	expectedSuccessRate := (1850.0 / 2050.0) * 100.0
	diff := res.Data.SuccessRate - expectedSuccessRate
	if diff < -0.1 || diff > 0.1 {
		t.Errorf("Expected SuccessRate ~%.2f%%, got %.2f%%", expectedSuccessRate, res.Data.SuccessRate)
	}

	// Avg latency = (10 + 20 + 0) / 3 = 10ms
	if res.Data.AvgLatencyMs != 10 {
		t.Errorf("Expected AvgLatencyMs = 10, got %d", res.Data.AvgLatencyMs)
	}
}

