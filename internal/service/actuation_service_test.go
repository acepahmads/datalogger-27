package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type mockModbusWriter struct {
	connected bool
	lastReq   *modbus.ModbusWriteRequest
	failWrite bool
}

func (m *mockModbusWriter) Connect(ctx context.Context) error {
	m.connected = true
	return nil
}

func (m *mockModbusWriter) IsConnected() bool {
	return m.connected
}

func (m *mockModbusWriter) WriteRegisters(ctx context.Context, req modbus.ModbusWriteRequest) (*modbus.ModbusWriteResponse, error) {
	m.lastReq = &req
	if m.failWrite {
		return nil, fmt.Errorf("simulated bus timeout")
	}
	return &modbus.ModbusWriteResponse{
		SlaveID:      req.SlaveID,
		FunctionCode: req.FunctionCode,
		Address:      req.Address,
		Value:        req.Value,
		ResponseTime: 12 * time.Millisecond,
		RequestADU:   []byte{req.SlaveID, req.FunctionCode, 0x00, byte(req.Address), byte(req.Value >> 8), byte(req.Value), 0xAA, 0xBB},
		ResponseADU:  []byte{req.SlaveID, req.FunctionCode, 0x00, byte(req.Address), byte(req.Value >> 8), byte(req.Value), 0xAA, 0xBB},
		CRCPassed:    true,
	}, nil
}

func setupActuatorTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test sqlite DB: %v", err)
	}

	err = db.AutoMigrate(
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.DeviceActuator{},
		&model.ActuationSchedule{},
		&model.ActuationLog{},
	)
	if err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}
	return db
}

func TestActuationService_CRUDAndDefaults(t *testing.T) {
	db := setupActuatorTestDB(t)
	actuatorRepo := repository.NewActuatorRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)

	mockWriter := &mockModbusWriter{connected: true}
	svc := service.NewActuationService(
		actuatorRepo,
		deviceRepo,
		func(dev *model.Device) (service.ModbusWriter, error) {
			return mockWriter, nil
		},
		nil,
	)

	// Create Device
	dev := &model.Device{
		DeviceCode: "DEV-TEST-01",
		DeviceName: "Test Modbus Station",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
		Connection: &model.DeviceConnection{
			Protocol: model.ProtocolModbusRTU,
			SlaveID:  1,
		},
	}
	if err := deviceRepo.Create(dev); err != nil {
		t.Fatalf("failed to create device: %v", err)
	}

	// 1. Create FC05 Coil Actuator (Test default 0xFF00 on-payload)
	act1 := &model.DeviceActuator{
		DeviceID:             dev.ID,
		ActuatorCode:         "PUMP-01",
		ActuatorName:         "Water Sampling Pump 01",
		ControlType:          model.ControlTypeCoil,
		SlaveID:              1,
		RegisterAddress:      10,
		SafetyTimeoutSeconds: 300,
		SafetyWatchdogActive: true,
		Enabled:              true,
	}
	if err := svc.CreateActuator(act1); err != nil {
		t.Fatalf("CreateActuator failed: %v", err)
	}
	if act1.OnPayload != 0xFF00 {
		t.Errorf("Expected FC05 OnPayload to default to 0xFF00, got 0x%04X", act1.OnPayload)
	}
	if act1.CurrentState != model.ActuatorStateOff {
		t.Errorf("Expected initial state to be OFF, got %s", act1.CurrentState)
	}

	// 2. Actuate Turn ON
	ctx := context.Background()
	logEntry, err := svc.Actuate(ctx, act1.ID, model.ActionTurnOn, "TEST_SUITE", nil)
	if err != nil {
		t.Fatalf("Actuate ON failed: %v", err)
	}
	if logEntry.Status != model.ActuationStatusSuccess {
		t.Errorf("Expected log status SUCCESS, got %s", logEntry.Status)
	}
	if mockWriter.lastReq.FunctionCode != modbus.FunctionWriteSingleCoil {
		t.Errorf("Expected FunctionCode 0x05, got 0x%02X", mockWriter.lastReq.FunctionCode)
	}
	if mockWriter.lastReq.Value != 0xFF00 {
		t.Errorf("Expected Value 0xFF00, got 0x%04X", mockWriter.lastReq.Value)
	}

	// Verify state in DB
	refreshed, err := svc.GetActuator(act1.ID)
	if err != nil {
		t.Fatalf("GetActuator failed: %v", err)
	}
	if refreshed.CurrentState != model.ActuatorStateOn {
		t.Errorf("Expected refreshed state ON, got %s", refreshed.CurrentState)
	}
	if refreshed.AutoOffDueAt == nil {
		t.Errorf("Expected AutoOffDueAt to be populated for safety watchdog")
	}

	// 3. Actuate Turn OFF
	logOff, err := svc.Actuate(ctx, act1.ID, model.ActionTurnOff, "TEST_SUITE", nil)
	if err != nil {
		t.Fatalf("Actuate OFF failed: %v", err)
	}
	if logOff.Status != model.ActuationStatusSuccess {
		t.Errorf("Expected log status SUCCESS, got %s", logOff.Status)
	}
	if mockWriter.lastReq.Value != 0x0000 {
		t.Errorf("Expected Value 0x0000, got 0x%04X", mockWriter.lastReq.Value)
	}

	refreshedOff, _ := svc.GetActuator(act1.ID)
	if refreshedOff.CurrentState != model.ActuatorStateOff {
		t.Errorf("Expected refreshed state OFF, got %s", refreshedOff.CurrentState)
	}
	if refreshedOff.AutoOffDueAt != nil {
		t.Errorf("Expected AutoOffDueAt to be nil when turned OFF")
	}

	// 4. Test Write Verification
	testRes, err := svc.TestWrite(ctx, service.ActuatorTestWriteRequest{
		ActuatorID: act1.ID,
		Action:     "ON",
	})
	if err != nil {
		t.Fatalf("TestWrite failed: %v", err)
	}
	if !testRes.Success {
		t.Errorf("Expected TestWrite to succeed")
	}
	if testRes.RawRequestHex == "" || testRes.RawResponseHex == "" {
		t.Errorf("Expected non-empty hex traces in test result")
	}
}
