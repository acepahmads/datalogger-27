package service_test

import (
	"fmt"
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
		Scale:         0.1,
		Offset:        0.0,
	}

	param, err := devService.CreateParameter(dev.ID, pReq, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to create parameter: %v", err)
	}
	if param.ID == 0 || param.ParameterCode != "TEMP_01" {
		t.Errorf("Expected valid parameter, got ID %d, Code %s", param.ID, param.ParameterCode)
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
	_, updateErr := devService.UpdateParameter(dev.ID, param.ID, &service.UpdateParameterRequest{
		ParameterName: &newName,
		Scale:         &newScale,
	}, "admin", "127.0.0.1", "TestAgent")
	if updateErr != nil {
		t.Fatalf("Failed to update parameter: %v", updateErr)
	}

	updatedParam, _ := devService.GetParameter(dev.ID, param.ID)
	if updatedParam.ParameterName != "Ambient Temperature Chamber" || updatedParam.Scale != 0.5 {
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

