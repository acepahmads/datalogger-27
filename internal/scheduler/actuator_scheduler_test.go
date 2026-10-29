package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/scheduler"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type recordingWriter struct {
	mu       sync.Mutex
	commands []modbus.ModbusWriteRequest
}

func (w *recordingWriter) Connect(ctx context.Context) error { return nil }
func (w *recordingWriter) IsConnected() bool                 { return true }
func (w *recordingWriter) WriteRegisters(ctx context.Context, req modbus.ModbusWriteRequest) (*modbus.ModbusWriteResponse, error) {
	w.mu.Lock()
	w.commands = append(w.commands, req)
	w.mu.Unlock()
	return &modbus.ModbusWriteResponse{
		SlaveID:      req.SlaveID,
		FunctionCode: req.FunctionCode,
		Address:      req.Address,
		Value:        req.Value,
		ResponseTime: 5 * time.Millisecond,
		CRCPassed:    true,
	}, nil
}

func TestActuatorScheduler_WatchdogAndRecurringCycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite DB: %v", err)
	}

	_ = db.AutoMigrate(
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.DeviceActuator{},
		&model.ActuationSchedule{},
		&model.ActuationLog{},
	)

	actuatorRepo := repository.NewActuatorRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)

	writer := &recordingWriter{}
	actSvc := service.NewActuationService(
		actuatorRepo,
		deviceRepo,
		func(dev *model.Device) (service.ModbusWriter, error) {
			return writer, nil
		},
		nil,
	)

	dev := &model.Device{
		DeviceCode: "DEV-AQMS-01",
		DeviceName: "AQMS Master Station",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
		Connection: &model.DeviceConnection{Protocol: model.ProtocolModbusRTU},
	}
	_ = deviceRepo.Create(dev)

	// Actuator that is currently ON with an expired watchdog timer (auto-off due 10s ago)
	past := time.Now().UTC().Add(-10 * time.Second)
	act := &model.DeviceActuator{
		DeviceID:             dev.ID,
		ActuatorCode:         "PURGE-VALVE",
		ActuatorName:         "Purge Air Valve",
		ControlType:          model.ControlTypeCoil,
		RegisterAddress:      5,
		CurrentState:         model.ActuatorStateOn,
		AutoOffDueAt:         &past,
		SafetyWatchdogActive: true,
		Enabled:              true,
	}
	_ = actuatorRepo.CreateActuator(act)

	schedWorker := scheduler.NewActuatorScheduler(actuatorRepo, actSvc)
	schedWorker.SetCheckInterval(50 * time.Millisecond)
	schedWorker.Start()

	// Wait 1-2 seconds for scheduler tick
	time.Sleep(1200 * time.Millisecond)
	schedWorker.Stop()

	// Verify that the watchdog automatically turned the valve OFF
	refreshed, _ := actuatorRepo.GetActuatorByID(act.ID)
	if refreshed.CurrentState != model.ActuatorStateOff {
		t.Errorf("Expected watchdog to transition actuator to OFF, got %s", refreshed.CurrentState)
	}

	writer.mu.Lock()
	defer writer.mu.Unlock()
	if len(writer.commands) == 0 {
		t.Fatalf("Expected at least one Modbus write command dispatched by watchdog")
	}
	lastCmd := writer.commands[len(writer.commands)-1]
	if lastCmd.Value != 0x0000 {
		t.Errorf("Expected watchdog write payload 0x0000 (OFF), got 0x%04X", lastCmd.Value)
	}
}
