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

func TestCalculateNextScheduleRun_CustomTimingModes(t *testing.T) {
	refTime := time.Date(2026, 10, 29, 13, 20, 0, 0, time.UTC) // Thursday 13:20

	// 1. Hourly at minute 45 -> should be today 13:45
	schedHourly := &model.ActuationSchedule{
		ScheduleType: model.ScheduleTypeHourlyMinute,
		MinuteOfHour: "45",
		DaysOfWeek:   "1,2,3,4,5,6,7",
	}
	next1 := service.CalculateNextScheduleRun(schedHourly, refTime)
	expected1 := time.Date(2026, 10, 29, 13, 45, 0, 0, time.UTC)
	if !next1.Equal(expected1) {
		t.Errorf("HourlyMinute expected %v, got %v", expected1, next1)
	}

	// 1b. If reference time is 13:45, next minute 45 should wrap to 14:45
	refAt45 := time.Date(2026, 10, 29, 13, 45, 0, 0, time.UTC)
	next1b := service.CalculateNextScheduleRun(schedHourly, refAt45)
	expected1b := time.Date(2026, 10, 29, 14, 45, 0, 0, time.UTC)
	if !next1b.Equal(expected1b) {
		t.Errorf("HourlyMinute wrap expected %v, got %v", expected1b, next1b)
	}

	// 1c. Multi-minute: "15, 45" from 13:20 -> should choose 13:45, then 14:15
	schedMulti := &model.ActuationSchedule{
		ScheduleType: model.ScheduleTypeHourlyMinute,
		MinuteOfHour: "15, 45",
		DaysOfWeek:   "1,2,3,4,5,6,7",
	}
	nextMulti1 := service.CalculateNextScheduleRun(schedMulti, refTime)
	if !nextMulti1.Equal(expected1) {
		t.Errorf("MultiMinute expected %v, got %v", expected1, nextMulti1)
	}
	nextMulti2 := service.CalculateNextScheduleRun(schedMulti, expected1)
	expectedMulti2 := time.Date(2026, 10, 29, 14, 15, 0, 0, time.UTC)
	if !nextMulti2.Equal(expectedMulti2) {
		t.Errorf("MultiMinute next expected %v, got %v", expectedMulti2, nextMulti2)
	}

	// 2. Time of Day: "16:30" from 13:20 -> should be today 16:30
	schedTimeOfDay := &model.ActuationSchedule{
		ScheduleType: model.ScheduleTypeTimeOfDay,
		StartTimeStr: "16:30",
		DaysOfWeek:   "1,2,3,4,5,6,7",
	}
	next2 := service.CalculateNextScheduleRun(schedTimeOfDay, refTime)
	expected2 := time.Date(2026, 10, 29, 16, 30, 0, 0, time.UTC)
	if !next2.Equal(expected2) {
		t.Errorf("TimeOfDay expected %v, got %v", expected2, next2)
	}

	// 2b. Time of Day: "08:00" from 13:20 -> already passed today, should be tomorrow (Friday 10/30) 08:00
	schedPassed := &model.ActuationSchedule{
		ScheduleType: model.ScheduleTypeTimeOfDay,
		StartTimeStr: "08:00",
		DaysOfWeek:   "1,2,3,4,5,6,7",
	}
	next2b := service.CalculateNextScheduleRun(schedPassed, refTime)
	expected2b := time.Date(2026, 10, 30, 8, 0, 0, 0, time.UTC)
	if !next2b.Equal(expected2b) {
		t.Errorf("TimeOfDay tomorrow expected %v, got %v", expected2b, next2b)
	}

	// 3. Cron Expression: "45 * * * *" -> top of 45th minute every hour
	schedCron := &model.ActuationSchedule{
		ScheduleType:   model.ScheduleTypeCron,
		CronExpression: "45 * * * *",
		DaysOfWeek:     "1,2,3,4,5,6,7",
	}
	next3 := service.CalculateNextScheduleRun(schedCron, refTime)
	if !next3.Equal(expected1) {
		t.Errorf("Cron expected %v, got %v", expected1, next3)
	}
}
