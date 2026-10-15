package service_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"datalogger/internal/database"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupDeviceTestDB(t *testing.T) (*gorm.DB, *service.DeviceService, *repository.SystemRepository) {
	dsn := fmt.Sprintf("file:mem_dev_srv_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	devRepo := repository.NewDeviceRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	devService := service.NewDeviceService(devRepo, sysRepo)

	return db, devService, sysRepo
}

func TestDeviceCreateAndDuplicateCode(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	// 1. Create valid device
	req := &service.CreateDeviceRequest{
		DeviceCode:   "TEST-DEV-01",
		DeviceName:   "Test Power Meter",
		DeviceType:   "MODBUS_TCP",
		Manufacturer: "Schneider",
		Model:        "PM5000",
		Status:       model.DeviceStatusActive,
		Connection: &service.ConnectionConfigDTO{
			Protocol: model.ProtocolModbusTCP,
			Host:     "192.168.1.50",
			Port:     502,
		},
	}

	dev, err := devService.CreateDevice(req, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	if dev.ID == 0 {
		t.Errorf("Expected valid ID, got 0")
	}
	if dev.DeviceCode != "TEST-DEV-01" {
		t.Errorf("Expected TEST-DEV-01, got %s", dev.DeviceCode)
	}
	if dev.Connection == nil || dev.Connection.Host != "192.168.1.50" {
		t.Errorf("Expected connection host 192.168.1.50")
	}

	// 2. Test duplicate device_code rejection
	dupReq := &service.CreateDeviceRequest{
		DeviceCode: "TEST-DEV-01",
		DeviceName: "Duplicate Meter",
		DeviceType: "MODBUS_TCP",
	}

	_, dupErr := devService.CreateDevice(dupReq, "admin", "127.0.0.1", "TestAgent")
	if dupErr == nil {
		t.Fatalf("Expected duplicate device_code to fail, but it succeeded")
	}
}

func TestDeviceUpdateAndSoftDelete(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	req := &service.CreateDeviceRequest{
		DeviceCode: "TEST-DEV-02",
		DeviceName: "Flow Meter Initial",
		DeviceType: "MODBUS_RTU",
		Status:     model.DeviceStatusActive,
	}

	dev, err := devService.CreateDevice(req, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	// Update device name and status
	newName := "Flow Meter Updated"
	newStatus := model.DeviceStatusMaintenance
	updateReq := &service.UpdateDeviceRequest{
		DeviceName: &newName,
		Status:     &newStatus,
	}

	updated, err := devService.UpdateDevice(dev.ID, updateReq, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to update device: %v", err)
	}

	if updated.DeviceName != "Flow Meter Updated" {
		t.Errorf("Expected updated name, got %s", updated.DeviceName)
	}
	if updated.Status != model.DeviceStatusMaintenance {
		t.Errorf("Expected MAINTENANCE status, got %s", updated.Status)
	}

	// Soft delete device
	if err := devService.DeleteDevice(dev.ID, "admin", "127.0.0.1", "TestAgent"); err != nil {
		t.Fatalf("Failed to delete device: %v", err)
	}

	// Device should no longer be fetched by regular query
	_, getErr := devService.GetDeviceByID(dev.ID)
	if getErr == nil {
		t.Errorf("Expected error fetching soft-deleted device, but found it")
	}
}

func TestDeviceListSearchAndFilter(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	// Seed 3 devices
	d1 := &service.CreateDeviceRequest{DeviceCode: "DEV-A", DeviceName: "Boiler Alpha", DeviceType: "MODBUS_TCP", Location: "Building 1", Status: model.DeviceStatusActive}
	d2 := &service.CreateDeviceRequest{DeviceCode: "DEV-B", DeviceName: "Chiller Beta", DeviceType: "MODBUS_RTU", Location: "Building 2", Status: model.DeviceStatusInactive}
	d3 := &service.CreateDeviceRequest{DeviceCode: "DEV-C", DeviceName: "Generator Gamma", DeviceType: "MODBUS_TCP", Location: "Building 1", Status: model.DeviceStatusActive}

	_, _ = devService.CreateDevice(d1, "admin", "127.0.0.1", "TestAgent")
	_, _ = devService.CreateDevice(d2, "admin", "127.0.0.1", "TestAgent")
	_, _ = devService.CreateDevice(d3, "admin", "127.0.0.1", "TestAgent")

	// Search "Alpha"
	items, total, err := devService.ListDevices(repository.DeviceFilterParams{Search: "Alpha"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if total != 1 || items[0].DeviceCode != "DEV-A" {
		t.Errorf("Expected 1 result DEV-A, got %d", total)
	}

	// Filter status "ACTIVE"
	_, activeTotal, err := devService.ListDevices(repository.DeviceFilterParams{Status: "ACTIVE"})
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}
	if activeTotal != 2 {
		t.Errorf("Expected 2 active devices, got %d", activeTotal)
	}

	// Pagination
	pagedItems, pagedTotal, err := devService.ListDevices(repository.DeviceFilterParams{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("Pagination failed: %v", err)
	}
	if pagedTotal != 3 || len(pagedItems) != 2 {
		t.Errorf("Expected 2 items out of 3, got %d (len %d)", pagedTotal, len(pagedItems))
	}
}

func TestDeviceValidation(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	// Missing code
	_, err := devService.CreateDevice(&service.CreateDeviceRequest{DeviceName: "Test"}, "admin", "127.0.0.1", "")
	if err == nil {
		t.Errorf("Expected error for missing device_code")
	}

	// Missing name
	_, err = devService.CreateDevice(&service.CreateDeviceRequest{DeviceCode: "DEV-X"}, "admin", "127.0.0.1", "")
	if err == nil {
		t.Errorf("Expected error for missing device_name")
	}

	// Invalid latitude
	badLat := 120.0
	_, err = devService.CreateDevice(&service.CreateDeviceRequest{DeviceCode: "DEV-LAT", DeviceName: "Bad Lat", Latitude: &badLat}, "admin", "127.0.0.1", "")
	if err == nil {
		t.Errorf("Expected error for invalid latitude")
	}
}

func TestDeviceHealthFoundation(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	dev, _ := devService.CreateDevice(&service.CreateDeviceRequest{DeviceCode: "HLTH-01", DeviceName: "Health Test Node"}, "admin", "127.0.0.1", "")

	// Set online
	if err := devService.SetOnline(dev.ID); err != nil {
		t.Fatalf("SetOnline failed: %v", err)
	}

	updated, _ := devService.GetDeviceByID(dev.ID)
	if updated.ConnectionStatus != model.DeviceConnOnline {
		t.Errorf("Expected ONLINE, got %s", updated.ConnectionStatus)
	}
	if updated.LastSeenAt == nil {
		t.Errorf("Expected non-nil LastSeenAt")
	}

	// Update last data
	if err := devService.UpdateLastData(dev.ID); err != nil {
		t.Fatalf("UpdateLastData failed: %v", err)
	}
	updated, _ = devService.GetDeviceByID(dev.ID)
	if updated.LastDataAt == nil {
		t.Errorf("Expected non-nil LastDataAt")
	}

	// Set connection error
	if err := devService.SetConnectionError(dev.ID, "Connection timed out"); err != nil {
		t.Fatalf("SetConnectionError failed: %v", err)
	}
	updated, _ = devService.GetDeviceByID(dev.ID)
	if updated.ConnectionStatus != model.DeviceConnError {
		t.Errorf("Expected ERROR, got %s", updated.ConnectionStatus)
	}
}

func TestParameterCRUDAndUniqueness(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	dev, _ := devService.CreateDevice(&service.CreateDeviceRequest{DeviceCode: "DEV-PARAM", DeviceName: "Param Node"}, "admin", "127.0.0.1", "")

	// 1. Create valid parameter
	pReq := &service.CreateParameterRequest{
		ParameterCode: "TEMP_01",
		ParameterName: "Temperature Sensor 1",
		DataType:      model.DataTypeFloat32,
		Unit:          "°C",
		Icon:          "thermometer",
		Scale:         0.1,
		Offset:        0.0,
	}

	param, err := devService.CreateParameter(dev.ID, pReq, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to create parameter: %v", err)
	}
	if param.ID == 0 || param.ParameterCode != "TEMP_01" || param.Icon != "thermometer" {
		t.Errorf("Expected valid parameter with icon, got ID %d, Code %s, Icon %s", param.ID, param.ParameterCode, param.Icon)
	}

	// 2. Reject duplicate parameter_code for SAME device
	_, dupErr := devService.CreateParameter(dev.ID, pReq, "admin", "127.0.0.1", "TestAgent")
	if dupErr == nil {
		t.Fatalf("Expected duplicate parameter_code on same device to fail, but it succeeded")
	}

	// 3. Different device CAN have the same parameter_code
	dev2, _ := devService.CreateDevice(&service.CreateDeviceRequest{DeviceCode: "DEV-PARAM-2", DeviceName: "Param Node 2"}, "admin", "127.0.0.1", "")
	_, errDev2 := devService.CreateParameter(dev2.ID, pReq, "admin", "127.0.0.1", "TestAgent")
	if errDev2 != nil {
		t.Fatalf("Expected same parameter code on different device to succeed, got %v", errDev2)
	}

	// 4. Update parameter
	newName := "Ambient Temperature Chamber"
	newScale := 0.5
	newIcon := "droplet"
	_, updateErr := devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		ParameterName: &newName,
		Scale:         &newScale,
		Icon:          &newIcon,
	}, "admin", "127.0.0.1", "TestAgent")
	if updateErr != nil {
		t.Fatalf("Failed to update parameter: %v", updateErr)
	}

	updatedParam, _ := devService.GetParameter(dev.ID, param.ID)
	if updatedParam.ParameterName != "Ambient Temperature Chamber" || updatedParam.Scale != 0.5 || updatedParam.Icon != "droplet" {
		t.Errorf("Parameter update failed: %+v", updatedParam)
	}

	// 5. Delete parameter
	if err := devService.DeleteParameter(dev.ID, param.ID, "admin", "127.0.0.1", "TestAgent"); err != nil {
		t.Fatalf("Failed to delete parameter: %v", err)
	}
	_, getDeletedErr := devService.GetParameter(dev.ID, param.ID)
	if getDeletedErr == nil {
		t.Errorf("Expected error fetching deleted parameter, but found it")
	}
}

func TestDeviceAuditTrail(t *testing.T) {
	_, devService, sysRepo := setupDeviceTestDB(t)

	// Create device -> should log CREATE_DEVICE
	dev, err := devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "AUDIT-01",
		DeviceName: "Audit Test Meter",
	}, "lead_engineer", "192.168.1.10", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("Create device failed: %v", err)
	}

	// Update device -> should log UPDATE_DEVICE
	newName := "Audit Test Meter Updated"
	_, _ = devService.UpdateDevice(dev.ID, &service.UpdateDeviceRequest{DeviceName: &newName}, "lead_engineer", "192.168.1.10", "Mozilla/5.0")

	// Delete device -> should log DELETE_DEVICE
	_ = devService.DeleteDevice(dev.ID, "lead_engineer", "192.168.1.10", "Mozilla/5.0")

	// Query audit trail
	trails, err := sysRepo.GetAuditTrails(50)
	if err != nil {
		t.Fatalf("Failed to fetch audit trails: %v", err)
	}

	hasCreate := false
	hasUpdate := false
	hasDelete := false
	for _, trail := range trails {
		if trail.Action == "CREATE_DEVICE" {
			hasCreate = true
		}
		if trail.Action == "UPDATE_DEVICE" {
			hasUpdate = true
		}
		if trail.Action == "DELETE_DEVICE" {
			hasDelete = true
		}
	}

	if !hasCreate {
		t.Errorf("Expected CREATE_DEVICE in audit trail")
	}
	if !hasUpdate {
		t.Errorf("Expected UPDATE_DEVICE in audit trail")
	}
	if !hasDelete {
		t.Errorf("Expected DELETE_DEVICE in audit trail")
	}
}

func TestParameterFormulaAndHoldConfig(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	dev, _ := devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "DEV-FORMULA-01",
		DeviceName: "Formula Node",
	}, "admin", "127.0.0.1", "")

	// 1. Invalid formula should fail validation
	_, errInvalid := devService.CreateParameter(dev.ID, &service.CreateParameterRequest{
		ParameterCode: "BAD_PARAM",
		ParameterName: "Bad Formula Param",
		Formula:       "x +* 2",
	}, "admin", "127.0.0.1", "")
	if errInvalid == nil {
		t.Errorf("Expected invalid formula syntax to be rejected, but succeeded")
	}

	// 2. Valid formula and hold configuration
	holdEnabled := true
	pReq := &service.CreateParameterRequest{
		ParameterCode:        "TEMP_F",
		ParameterName:        "Temperature Fahrenheit",
		DataType:             model.DataTypeFloat32,
		Unit:                 "°F",
		Formula:              "x * 1.8 + 32",
		HoldLastValueEnabled: &holdEnabled,
		HoldLastValueSeconds: 120,
	}

	param, err := devService.CreateParameter(dev.ID, pReq, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("Failed to create parameter with formula: %v", err)
	}
	if param.Formula != "x * 1.8 + 32" {
		t.Errorf("Expected formula 'x * 1.8 + 32', got '%s'", param.Formula)
	}
	if !param.HoldLastValueEnabled {
		t.Errorf("Expected HoldLastValueEnabled true")
	}
	if param.HoldLastValueSeconds != 120 {
		t.Errorf("Expected HoldLastValueSeconds 120, got %d", param.HoldLastValueSeconds)
	}

	// 3. Update formula to another expression
	newFormula := "(x - 4) * 6.25"
	newSeconds := 60
	updated, err := devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		Formula:              &newFormula,
		HoldLastValueSeconds: &newSeconds,
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("Failed to update parameter formula: %v", err)
	}
	if updated.Formula != newFormula {
		t.Errorf("Expected updated formula '%s', got '%s'", newFormula, updated.Formula)
	}
	if updated.HoldLastValueSeconds != 60 {
		t.Errorf("Expected updated HoldLastValueSeconds 60, got %d", updated.HoldLastValueSeconds)
	}
}

func TestParameterJSONFlexibleUnmarshal(t *testing.T) {
	// Test unmarshaling empty strings and stringified numbers
	rawJSON := []byte(`{
		"min_value": "",
		"max_value": "150.5",
		"warning_low": "20",
		"warning_high": null,
		"scale": "1.25",
		"offset": "0.5"
	}`)

	var req service.UpdateParameterRequest
	if err := json.Unmarshal(rawJSON, &req); err != nil {
		t.Fatalf("Failed to unmarshal flexible JSON: %v", err)
	}

	if !req.ClearMinValue || req.MinValue != nil {
		t.Errorf("Expected ClearMinValue true and MinValue nil, got Clear=%v, MinValue=%v", req.ClearMinValue, req.MinValue)
	}
	if req.MaxValue == nil || *req.MaxValue != 150.5 {
		t.Errorf("Expected MaxValue 150.5, got %v", req.MaxValue)
	}
	if req.WarningLow == nil || *req.WarningLow != 20.0 {
		t.Errorf("Expected WarningLow 20.0, got %v", req.WarningLow)
	}
	if !req.ClearWarningHigh || req.WarningHigh != nil {
		t.Errorf("Expected ClearWarningHigh true and WarningHigh nil, got Clear=%v, WarningHigh=%v", req.ClearWarningHigh, req.WarningHigh)
	}
	if req.Scale == nil || *req.Scale != 1.25 {
		t.Errorf("Expected Scale 1.25, got %v", req.Scale)
	}
	if req.Offset == nil || *req.Offset != 0.5 {
		t.Errorf("Expected Offset 0.5, got %v", req.Offset)
	}
}

func TestUpdateParameterLimitsAndClear(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	dev, err := devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "DEV-LIMITS-01",
		DeviceName: "Limit Validation Test Device",
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	minVal := 0.0
	maxVal := 100.0
	warnLow := 10.0
	warnHigh := 50.0
	param, err := devService.CreateParameter(dev.ID, &service.CreateParameterRequest{
		ParameterCode:   "TEMP_CH1",
		ParameterName:   "Chamber 1 Temperature",
		DataType:        model.DataTypeFloat32,
		MinValue:        &minVal,
		MaxValue:        &maxVal,
		WarningLow:      &warnLow,
		WarningHigh:     &warnHigh,
		Scale:           1.0,
		RegisterAddress: 4,
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("CreateParameter failed: %v", err)
	}

	// 1. Validation check: warning_low > warning_high should fail
	invalidWarnLow := 60.0
	_, err = devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		WarningLow: &invalidWarnLow,
	}, "admin", "127.0.0.1", "")
	if err == nil {
		t.Fatalf("Expected error when warning_low (60) > existing warning_high (50), got nil")
	}

	// 2. Validation check: min_value > max_value should fail
	invalidMinVal := 120.0
	_, err = devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		MinValue: &invalidMinVal,
	}, "admin", "127.0.0.1", "")
	if err == nil {
		t.Fatalf("Expected error when min_value (120) > existing max_value (100), got nil")
	}

	// 3. Clear warning_high and update warning_low to 60.0 -> should succeed because warning_high is cleared!
	newWarnLow := 60.0
	updated, err := devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		WarningLow:       &newWarnLow,
		ClearWarningHigh: true,
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("UpdateParameter with ClearWarningHigh failed: %v", err)
	}
	if updated.WarningHigh != nil {
		t.Errorf("Expected WarningHigh to be nil after clear, got %v", *updated.WarningHigh)
	}
	if updated.WarningLow == nil || *updated.WarningLow != 60.0 {
		t.Errorf("Expected WarningLow to be 60.0, got %v", updated.WarningLow)
	}

	// 4. Clear all limits
	updated2, err := devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		ClearMinValue:    true,
		ClearMaxValue:    true,
		ClearWarningLow:  true,
		ClearWarningHigh: true,
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("UpdateParameter clearing all limits failed: %v", err)
	}
	if updated2.MinValue != nil || updated2.MaxValue != nil || updated2.WarningLow != nil || updated2.WarningHigh != nil {
		t.Errorf("Expected all limits to be nil, got: min=%v, max=%v, wlow=%v, whigh=%v",
			updated2.MinValue, updated2.MaxValue, updated2.WarningLow, updated2.WarningHigh)
	}
}

func TestUpdateParameterZeroValueAlarmPolicy(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	dev, err := devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "DEV-ZERO-POLICY-01",
		DeviceName: "Zero Policy Device",
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	param, err := devService.CreateParameter(dev.ID, &service.CreateParameterRequest{
		ParameterCode: "WIND_SPEED",
		ParameterName: "Wind Speed Sensor",
		DataType:      model.DataTypeFloat32,
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("CreateParameter failed: %v", err)
	}

	if param.ZeroValueAlarmPolicy != model.ZeroValuePolicyNormal {
		t.Errorf("Expected default policy NORMAL, got %v", param.ZeroValueAlarmPolicy)
	}

	// Update via JSON payload as sent by frontend
	rawJSON := []byte(`{
		"parameter_code": "WIND_SPEED",
		"parameter_name": "Wind Speed Sensor",
		"zero_value_alarm_policy": "SUPPRESS_THRESHOLD"
	}`)
	var updateReq service.UpdateParameterRequest
	if err := json.Unmarshal(rawJSON, &updateReq); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	updated, err := devService.UpdateParameter(dev.ID, param.ID, &updateReq, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("UpdateParameter failed: %v", err)
	}

	if updated.ZeroValueAlarmPolicy != model.ZeroValuePolicySuppressThreshold {
		t.Errorf("Expected updated policy SUPPRESS_THRESHOLD, got %v", updated.ZeroValueAlarmPolicy)
	}
}

func TestSingleProtocolSelectionAndSync(t *testing.T) {
	_, devService, _ := setupDeviceTestDB(t)

	// 1. Create device providing only Connection.Protocol = MODBUS_RTU (DeviceType empty)
	req := &service.CreateDeviceRequest{
		DeviceCode: "PROTO-DEV-01",
		DeviceName: "RTU Sensor",
		DeviceType: "", // UI now sends empty or relies on connection.protocol
		Connection: &service.ConnectionConfigDTO{
			Protocol:        model.ProtocolModbusRTU,
			SerialPort:      "/dev/ttyUSB0",
			BaudRate:        9600,
			DataBits:        8,
			Parity:          "N",
			StopBits:        1,
			SlaveID:         1,
			Timeout:         1000,
			RetryCount:      3,
			PollingInterval: 1000,
		},
	}

	dev, err := devService.CreateDevice(req, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	// Verify DeviceType is automatically synchronized to Connection.Protocol
	if dev.DeviceType != string(model.ProtocolModbusRTU) {
		t.Errorf("Expected DeviceType to be synchronized to %s, got %s", model.ProtocolModbusRTU, dev.DeviceType)
	}
	if dev.Connection == nil || dev.Connection.Protocol != model.ProtocolModbusRTU {
		t.Fatalf("Expected Connection.Protocol to be %s", model.ProtocolModbusRTU)
	}
	if dev.Connection.SerialPort != "/dev/ttyUSB0" || dev.Connection.BaudRate != 9600 {
		t.Errorf("Expected serial connection parameters to be preserved, got port=%s baud=%d",
			dev.Connection.SerialPort, dev.Connection.BaudRate)
	}

	// 2. Normalization test: HTTP/REST should normalize to HTTP
	httpReq := &service.CreateDeviceRequest{
		DeviceCode: "PROTO-DEV-02",
		DeviceName: "REST Gateway",
		Connection: &service.ConnectionConfigDTO{
			Protocol: "HTTP/REST",
			Host:     "192.168.1.100",
			Port:     8080,
		},
	}
	httpDev, err := devService.CreateDevice(httpReq, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("CreateDevice with HTTP/REST failed: %v", err)
	}
	if httpDev.DeviceType != string(model.ProtocolHTTP) {
		t.Errorf("Expected DeviceType normalized to %s, got %s", model.ProtocolHTTP, httpDev.DeviceType)
	}
	if httpDev.Connection.Protocol != model.ProtocolHTTP {
		t.Errorf("Expected Connection.Protocol normalized to %s, got %s", model.ProtocolHTTP, httpDev.Connection.Protocol)
	}

	// 3. Update device changing protocol to MODBUS_TCP
	updateReq := &service.UpdateDeviceRequest{
		Connection: &service.ConnectionConfigDTO{
			Protocol:        model.ProtocolModbusTCP,
			Host:            "10.0.0.50",
			Port:            502,
			SlaveID:         1,
			Timeout:         500,
			RetryCount:      2,
			PollingInterval: 2000,
		},
	}
	updated, err := devService.UpdateDevice(dev.ID, updateReq, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("UpdateDevice failed: %v", err)
	}

	if updated.DeviceType != string(model.ProtocolModbusTCP) {
		t.Errorf("Expected updated DeviceType to be %s, got %s", model.ProtocolModbusTCP, updated.DeviceType)
	}
	if updated.Connection == nil || updated.Connection.Protocol != model.ProtocolModbusTCP {
		t.Fatalf("Expected updated Connection.Protocol to be %s", model.ProtocolModbusTCP)
	}
	if updated.Connection.Host != "10.0.0.50" || updated.Connection.Port != 502 {
		t.Errorf("Expected TCP host/port to be updated, got host=%s port=%d", updated.Connection.Host, updated.Connection.Port)
	}

	// 4. Reopen/get device and verify persistence
	retrieved, err := devService.GetDeviceByID(dev.ID)
	if err != nil {
		t.Fatalf("GetDeviceByID failed: %v", err)
	}
	if retrieved.DeviceType != string(model.ProtocolModbusTCP) {
		t.Errorf("Expected retrieved DeviceType to be %s, got %s", model.ProtocolModbusTCP, retrieved.DeviceType)
	}
	if retrieved.Connection == nil || retrieved.Connection.Protocol != model.ProtocolModbusTCP {
		t.Errorf("Expected retrieved Connection.Protocol to be %s", model.ProtocolModbusTCP)
	}
}

func TestCloneDevice(t *testing.T) {
	db, devService, _ := setupDeviceTestDB(t)

	// 1. Create source device
	srcDev, err := devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "AQMS-SRC-01",
		DeviceName: "Source AQMS Station",
		DeviceType: "MODBUS_RTU",
		Connection: &service.ConnectionConfigDTO{
			Protocol:        model.ProtocolModbusRTU,
			SerialPort:      "COM3",
			BaudRate:        9600,
			DataBits:        8,
			Parity:          "N",
			StopBits:        1,
			SlaveID:         1,
			Timeout:         1000,
			RetryCount:      3,
			PollingInterval: 2000,
		},
	}, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to create source device: %v", err)
	}

	// 2. Clone device with new unique code and SourceDeviceID
	clonedDev, err := devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode:     "AQMS-CLONE-02",
		DeviceName:     "Cloned AQMS Station",
		DeviceType:     srcDev.DeviceType,
		SourceDeviceID: &srcDev.ID,
		Connection: &service.ConnectionConfigDTO{
			Protocol:        model.ProtocolModbusRTU,
			SerialPort:      "COM3",
			BaudRate:        9600,
			DataBits:        8,
			Parity:          "N",
			StopBits:        1,
			SlaveID:         2, // distinct slave ID
			Timeout:         1000,
			RetryCount:      3,
			PollingInterval: 2000,
		},
	}, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to clone device: %v", err)
	}

	// Verify distinct identity and untouched source
	if clonedDev.ID == srcDev.ID {
		t.Fatalf("Expected cloned device to have new ID, got identical ID %d", clonedDev.ID)
	}
	if clonedDev.DeviceCode != "AQMS-CLONE-02" {
		t.Errorf("Expected cloned code AQMS-CLONE-02, got %s", clonedDev.DeviceCode)
	}

	freshSrc, err := devService.GetDeviceByID(srcDev.ID)
	if err != nil {
		t.Fatalf("Failed to fetch fresh source: %v", err)
	}
	if freshSrc.DeviceCode != "AQMS-SRC-01" {
		t.Errorf("Source device code modified! Got %s", freshSrc.DeviceCode)
	}

	// 3. Verify duplicate code rejection
	_, err = devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "AQMS-SRC-01",
		DeviceName: "Duplicate Code Attempt",
	}, "admin", "127.0.0.1", "TestAgent")
	if err == nil {
		t.Fatalf("Expected duplicate device code to be rejected, but succeeded")
	}

	// 4. Verify audit trail recorded CLONE_DEVICE
	var audits []model.AuditTrail
	err = db.Where("action = ?", "CLONE_DEVICE").Find(&audits).Error
	if err != nil {
		t.Fatalf("Failed to query audit logs: %v", err)
	}
	if len(audits) == 0 {
		t.Fatalf("Expected CLONE_DEVICE audit log, found 0")
	}
	foundCloneAudit := false
	for _, a := range audits {
		if a.Action == "CLONE_DEVICE" && strings.Contains(a.Details, fmt.Sprintf(`"device_id":%d`, clonedDev.ID)) {
			foundCloneAudit = true
			if !strings.Contains(a.Details, "cloned_from_device_id") {
				t.Errorf("Expected audit Details to contain cloned_from_device_id, got: %s", a.Details)
			}
		}
	}
	if !foundCloneAudit {
		t.Errorf("Did not find expected CLONE_DEVICE audit record for device %d", clonedDev.ID)
	}
}

func TestCloneParameter(t *testing.T) {
	db, devService, _ := setupDeviceTestDB(t)

	dev, err := devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "DEV-PARAM-CLONE-01",
		DeviceName: "Parameter Clone Test Device",
	}, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	minVal := 0.0
	maxVal := 500.0
	warnLow := 10.0
	warnHigh := 150.0
	scale := 0.1
	offset := 5.0

	// 1. Create source parameter
	srcParam, err := devService.CreateParameter(dev.ID, &service.CreateParameterRequest{
		ParameterCode:   "PM10_SRC",
		ParameterName:   "Particulate PM10",
		DataType:        model.DataTypeFloat32,
		Unit:            "µg/m³",
		RegisterAddress: 22,
		RegisterType:    "HOLDING_REGISTER",
		Scale:           scale,
		Offset:          offset,
		MinValue:        &minVal,
		MaxValue:        &maxVal,
		WarningLow:      &warnLow,
		WarningHigh:     &warnHigh,
	}, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("CreateParameter source failed: %v", err)
	}

	// 2. Clone parameter to PM25 on same device with SourceParameterID
	clonedParam, err := devService.CreateParameter(dev.ID, &service.CreateParameterRequest{
		ParameterCode:     "PM25_CLONE",
		ParameterName:     "Particulate PM2.5",
		DataType:          srcParam.DataType,
		Unit:              srcParam.Unit,
		RegisterAddress:   24, // adjusted register address
		RegisterType:      srcParam.RegisterType,
		Scale:             srcParam.Scale,
		Offset:            srcParam.Offset,
		MinValue:          srcParam.MinValue,
		MaxValue:          srcParam.MaxValue,
		WarningLow:        srcParam.WarningLow,
		WarningHigh:       srcParam.WarningHigh,
		SourceParameterID: &srcParam.ID,
	}, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("CreateParameter cloned failed: %v", err)
	}

	// Verify identity and fields
	if clonedParam.ID == srcParam.ID {
		t.Fatalf("Expected new parameter ID, got same as source: %d", clonedParam.ID)
	}
	if clonedParam.ParameterCode != "PM25_CLONE" {
		t.Errorf("Expected parameter code PM25_CLONE, got %s", clonedParam.ParameterCode)
	}
	if clonedParam.RegisterAddress != 24 {
		t.Errorf("Expected register address 24, got %d", clonedParam.RegisterAddress)
	}
	if clonedParam.Unit != "µg/m³" {
		t.Errorf("Expected unit µg/m³, got %s", clonedParam.Unit)
	}

	// 3. Verify duplicate code rejection on same device
	_, err = devService.CreateParameter(dev.ID, &service.CreateParameterRequest{
		ParameterCode: "PM10_SRC",
		ParameterName: "Duplicate Param Code",
		DataType:      model.DataTypeFloat32,
	}, "admin", "127.0.0.1", "TestAgent")
	if err == nil {
		t.Fatalf("Expected duplicate parameter code to fail, but succeeded")
	}

	// 4. Verify audit trail recorded CLONE_PARAMETER
	var audits []model.AuditTrail
	err = db.Where("action = ?", "CLONE_PARAMETER").Find(&audits).Error
	if err != nil {
		t.Fatalf("Failed to query audit logs: %v", err)
	}
	if len(audits) == 0 {
		t.Fatalf("Expected CLONE_PARAMETER audit log, found 0")
	}
	foundParamCloneAudit := false
	for _, a := range audits {
		if a.Action == "CLONE_PARAMETER" && strings.Contains(a.Details, fmt.Sprintf(`"parameter_id":%d`, clonedParam.ID)) {
			foundParamCloneAudit = true
			if !strings.Contains(a.Details, "cloned_from_parameter_id") {
				t.Errorf("Expected audit Details to contain cloned_from_parameter_id, got: %s", a.Details)
			}
		}
	}
	if !foundParamCloneAudit {
		t.Errorf("Did not find expected CLONE_PARAMETER audit record for parameter %d", clonedParam.ID)
	}
}

func TestUpdateParameterCategoryAssignment(t *testing.T) {
	db, devService, _ := setupDeviceTestDB(t)

	// Fetch seeded categories from migrations
	var cat1, cat2 model.ParameterCategory
	if err := db.Where("code = ?", "gas_pollutants").First(&cat1).Error; err != nil {
		t.Fatalf("Failed to find gas_pollutants category: %v", err)
	}
	if err := db.Where("code = ?", "particulate_matter").First(&cat2).Error; err != nil {
		t.Fatalf("Failed to find particulate_matter category: %v", err)
	}

	// Create device
	dev, err := devService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "DEV-CAT-01",
		DeviceName: "Category Test Device",
		DeviceType: "MODBUS_TCP",
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	// Create parameter with category 1
	param, err := devService.CreateParameter(dev.ID, &service.CreateParameterRequest{
		ParameterCode: "PM10",
		ParameterName: "PM 10",
		DataType:      model.DataTypeFloat32,
		CategoryID:    &cat1.ID,
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("Failed to create parameter: %v", err)
	}

	if param.CategoryID == nil || *param.CategoryID != cat1.ID {
		t.Fatalf("Expected category %d, got %v", cat1.ID, param.CategoryID)
	}

	// Step 1: Update to Category 2
	updated, err := devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		CategoryID: &cat2.ID,
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("Failed to update parameter category: %v", err)
	}
	if updated.CategoryID == nil || *updated.CategoryID != cat2.ID {
		t.Errorf("Step 1 return: Expected category %d, got %v", cat2.ID, updated.CategoryID)
	}

	// Fetch again from DB
	fetched, err := devService.GetParameter(dev.ID, param.ID)
	if err != nil {
		t.Fatalf("Failed to get parameter: %v", err)
	}
	if fetched.CategoryID == nil || *fetched.CategoryID != cat2.ID {
		t.Errorf("Step 1 DB persist: Expected category %d, got %v", cat2.ID, fetched.CategoryID)
	}
	if fetched.Category == nil || fetched.Category.ID != cat2.ID {
		t.Errorf("Step 1 Category struct: Expected Category ID %d, got %+v", cat2.ID, fetched.Category)
	}

	// Check device repo GetByID
	devWithParams, err := devService.GetDeviceByID(dev.ID)
	if err != nil {
		t.Fatalf("Failed to get device by id: %v", err)
	}
	if len(devWithParams.Parameters) == 0 {
		t.Fatalf("Expected parameters in devWithParams, got 0")
	}
	p0 := devWithParams.Parameters[0]
	if p0.CategoryID == nil || *p0.CategoryID != cat2.ID {
		t.Errorf("devWithParams Parameter CategoryID: Expected %d, got %v", cat2.ID, p0.CategoryID)
	}
	if p0.Category == nil || p0.Category.ID != cat2.ID {
		t.Errorf("devWithParams Parameter Category struct: Expected ID %d, got %+v", cat2.ID, p0.Category)
	}

	// Check device repo ListDevices (which OCC calls)
	listedDevices, _, err := devService.ListDevices(repository.DeviceFilterParams{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("Failed to list devices: %v", err)
	}
	var foundDev *model.Device
	for i := range listedDevices {
		if listedDevices[i].ID == dev.ID {
			foundDev = &listedDevices[i]
			break
		}
	}
	if foundDev == nil || len(foundDev.Parameters) == 0 {
		t.Fatalf("Expected foundDev with parameters from ListDevices")
	}
	lp0 := foundDev.Parameters[0]
	if lp0.CategoryID == nil || *lp0.CategoryID != cat2.ID {
		t.Errorf("ListDevices Parameter CategoryID: Expected %d, got %v", cat2.ID, lp0.CategoryID)
	}
	if lp0.Category == nil || lp0.Category.ID != cat2.ID {
		t.Errorf("ListDevices Parameter Category struct: Expected ID %d, got %+v", cat2.ID, lp0.Category)
	}

	// Step 2: Clear Category
	cleared, err := devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		ClearCategoryID: true,
	}, "admin", "127.0.0.1", "")
	if err != nil {
		t.Fatalf("Failed to clear parameter category: %v", err)
	}
	if cleared.CategoryID != nil {
		t.Errorf("Step 2 return: Expected nil category, got %v", cleared.CategoryID)
	}

	// Fetch again from DB
	fetchedCleared, err := devService.GetParameter(dev.ID, param.ID)
	if err != nil {
		t.Fatalf("Failed to get cleared parameter: %v", err)
	}
	if fetchedCleared.CategoryID != nil {
		t.Errorf("Step 2 DB persist: Expected nil category, got %v", fetchedCleared.CategoryID)
	}
}

