package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"
)

type ModbusWriter interface {
	WriteRegisters(ctx context.Context, req modbus.ModbusWriteRequest) (*modbus.ModbusWriteResponse, error)
	IsConnected() bool
	Connect(ctx context.Context) error
}

type AdapterProvider func(device *model.Device) (ModbusWriter, error)

type ActuatorTestWriteRequest struct {
	ActuatorID   uint    `json:"actuator_id"`
	DeviceID     uint    `json:"device_id"`
	ControlType  string  `json:"control_type"`  // e.g. "COIL_FC05", "HOLDING_REGISTER_FC06"
	SlaveID      byte    `json:"slave_id"`
	Address      uint16  `json:"address"`
	Value        uint16  `json:"value"`         // value to test write
	Action       string  `json:"action"`        // "ON", "OFF", "TEST"
}

type ActuationTestResult struct {
	Success        bool      `json:"success"`
	LatencyMs      int       `json:"latency_ms"`
	FunctionCode   byte      `json:"function_code"`
	SlaveID        byte      `json:"slave_id"`
	Address        uint16    `json:"address"`
	ValueWritten   uint16    `json:"value_written"`
	RawRequestHex  string    `json:"raw_request_hex"`
	RawResponseHex string    `json:"raw_response_hex"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

type ActuationService struct {
	actuatorRepo    *repository.ActuatorRepository
	deviceRepo      *repository.DeviceRepository
	adapterProvider AdapterProvider
	hub             *websocket.Hub
}

func NewActuationService(
	actuatorRepo *repository.ActuatorRepository,
	deviceRepo *repository.DeviceRepository,
	adapterProvider AdapterProvider,
	hub *websocket.Hub,
) *ActuationService {
	return &ActuationService{
		actuatorRepo:    actuatorRepo,
		deviceRepo:      deviceRepo,
		adapterProvider: adapterProvider,
		hub:             hub,
	}
}

// -------------------------------------------------------------
// CRUD Operations
// -------------------------------------------------------------

func (s *ActuationService) CreateActuator(actuator *model.DeviceActuator) error {
	if actuator.ActuatorCode == "" {
		return fmt.Errorf("actuator_code is required")
	}
	if actuator.ActuatorName == "" {
		return fmt.Errorf("actuator_name is required")
	}
	if actuator.DeviceID == 0 {
		return fmt.Errorf("device_id is required")
	}

	// Default payloads if unset
	if actuator.ControlType == model.ControlTypeCoil && actuator.OnPayload == 0 {
		actuator.OnPayload = 0xFF00
	} else if actuator.ControlType == model.ControlTypeHoldingRegister && actuator.OnPayload == 0 {
		actuator.OnPayload = 1
	}
	if actuator.SafetyTimeoutSeconds <= 0 {
		actuator.SafetyTimeoutSeconds = 900 // 15 mins default watchdog
	}
	actuator.CurrentState = model.ActuatorStateOff

	return s.actuatorRepo.CreateActuator(actuator)
}

func (s *ActuationService) UpdateActuator(actuator *model.DeviceActuator) error {
	existing, err := s.actuatorRepo.GetActuatorByID(actuator.ID)
	if err != nil {
		return fmt.Errorf("actuator not found: %w", err)
	}

	existing.ActuatorName = actuator.ActuatorName
	existing.ActuatorCode = actuator.ActuatorCode
	existing.Description = actuator.Description
	existing.Icon = actuator.Icon
	existing.ControlType = actuator.ControlType
	existing.SlaveID = actuator.SlaveID
	existing.RegisterAddress = actuator.RegisterAddress
	existing.OnPayload = actuator.OnPayload
	existing.OffPayload = actuator.OffPayload
	existing.SafetyTimeoutSeconds = actuator.SafetyTimeoutSeconds
	existing.SafetyWatchdogActive = actuator.SafetyWatchdogActive
	existing.FeedbackParameterID = actuator.FeedbackParameterID
	existing.Enabled = actuator.Enabled

	return s.actuatorRepo.UpdateActuator(existing)
}

func (s *ActuationService) DeleteActuator(id uint) error {
	return s.actuatorRepo.DeleteActuator(id)
}

func (s *ActuationService) GetActuator(id uint) (*model.DeviceActuator, error) {
	return s.actuatorRepo.GetActuatorByID(id)
}

func (s *ActuationService) ListActuatorsByDevice(deviceID uint) ([]model.DeviceActuator, error) {
	return s.actuatorRepo.ListActuatorsByDevice(deviceID)
}

func (s *ActuationService) ListAllActiveActuators() ([]model.DeviceActuator, error) {
	return s.actuatorRepo.ListAllActiveActuators()
}

// -------------------------------------------------------------
// Modbus Write Execution & Actuation
// -------------------------------------------------------------

func resolveFunctionCode(ct model.ControlType) byte {
	switch ct {
	case model.ControlTypeCoil:
		return modbus.FunctionWriteSingleCoil // 0x05
	case model.ControlTypeHoldingRegister:
		return modbus.FunctionWriteSingleRegister // 0x06
	case model.ControlTypeMultipleCoils:
		return modbus.FunctionWriteMultipleCoils // 0x0F
	case model.ControlTypeMultipleRegisters:
		return modbus.FunctionWriteMultipleRegisters // 0x10
	default:
		return modbus.FunctionWriteSingleCoil
	}
}

// TestWrite executes a transient Modbus write frame and reports roundtrip verification
func (s *ActuationService) TestWrite(ctx context.Context, req ActuatorTestWriteRequest) (*ActuationTestResult, error) {
	var deviceID uint = req.DeviceID
	var slaveID byte = req.SlaveID
	var address uint16 = req.Address
	var value uint16 = req.Value
	var fc byte

	if req.ActuatorID > 0 {
		actuator, err := s.actuatorRepo.GetActuatorByID(req.ActuatorID)
		if err != nil {
			return nil, fmt.Errorf("actuator %d not found: %w", req.ActuatorID, err)
		}
		deviceID = actuator.DeviceID
		slaveID = actuator.SlaveID
		address = actuator.RegisterAddress
		fc = resolveFunctionCode(actuator.ControlType)

		if req.Action == "ON" {
			value = actuator.OnPayload
			if actuator.ControlType == model.ControlTypeCoil && value == 0 {
				value = 0xFF00
			}
		} else if req.Action == "OFF" {
			value = actuator.OffPayload
		}
	} else {
		fc = resolveFunctionCode(model.ControlType(req.ControlType))
	}

	if deviceID == 0 {
		return nil, fmt.Errorf("device_id is required for Modbus test write")
	}

	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil || device == nil {
		return nil, fmt.Errorf("device %d not found: %w", deviceID, err)
	}

	adapter, err := s.adapterProvider(device)
	if err != nil {
		return nil, fmt.Errorf("failed to get communication adapter for device %s: %w", device.DeviceCode, err)
	}

	if !adapter.IsConnected() {
		connCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := adapter.Connect(connCtx); err != nil {
			return nil, fmt.Errorf("failed to connect to device %s: %w", device.DeviceCode, err)
		}
	}

	writeReq := modbus.ModbusWriteRequest{
		SlaveID:      slaveID,
		FunctionCode: fc,
		Address:      address,
		Value:        value,
	}

	writeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := adapter.WriteRegisters(writeCtx, writeReq)
	latency := int(time.Since(start).Milliseconds())

	result := &ActuationTestResult{
		Timestamp:    time.Now().UTC(),
		LatencyMs:    latency,
		FunctionCode: fc,
		SlaveID:      slaveID,
		Address:      address,
		ValueWritten: value,
	}

	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		if txErr, ok := err.(*modbus.ModbusTransactionError); ok {
			result.RawRequestHex = hex.EncodeToString(txErr.RequestADU)
			result.RawResponseHex = hex.EncodeToString(txErr.ResponseADU)
		}
		return result, nil
	}

	result.Success = true
	result.LatencyMs = int(resp.ResponseTime.Milliseconds())
	result.RawRequestHex = hex.EncodeToString(resp.RequestADU)
	result.RawResponseHex = hex.EncodeToString(resp.ResponseADU)

	return result, nil
}

// Actuate executes an ON or OFF command, updates device actuator state, logs audit entry, and broadcasts WS event
func (s *ActuationService) Actuate(ctx context.Context, actuatorID uint, action model.ActuationAction, triggeredBy string, scheduleID *uint) (*model.ActuationLog, error) {
	actuator, err := s.actuatorRepo.GetActuatorByID(actuatorID)
	if err != nil {
		return nil, fmt.Errorf("actuator %d not found: %w", actuatorID, err)
	}

	if !actuator.Enabled {
		return nil, fmt.Errorf("actuator %s (%s) is disabled", actuator.ActuatorName, actuator.ActuatorCode)
	}

	device, err := s.deviceRepo.GetByID(actuator.DeviceID)
	if err != nil || device == nil {
		return nil, fmt.Errorf("device %d not found for actuator %s", actuator.DeviceID, actuator.ActuatorCode)
	}

	var value uint16
	if action == model.ActionTurnOn {
		value = actuator.OnPayload
		if actuator.ControlType == model.ControlTypeCoil && value == 0 {
			value = 0xFF00
		}
	} else {
		value = actuator.OffPayload
	}

	fc := resolveFunctionCode(actuator.ControlType)

	adapter, err := s.adapterProvider(device)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain communication adapter for device %s: %w", device.DeviceCode, err)
	}

	if !adapter.IsConnected() {
		connCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		_ = adapter.Connect(connCtx)
		cancel()
	}

	writeReq := modbus.ModbusWriteRequest{
		SlaveID:      actuator.SlaveID,
		FunctionCode: fc,
		Address:      actuator.RegisterAddress,
		Value:        value,
	}

	writeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	start := time.Now()
	resp, writeErr := adapter.WriteRegisters(writeCtx, writeReq)
	latency := int(time.Since(start).Milliseconds())

	logEntry := &model.ActuationLog{
		ActuatorID:     actuatorID,
		ScheduleID:     scheduleID,
		DeviceID:       actuator.DeviceID,
		Action:         action,
		TriggeredBy:    triggeredBy,
		FunctionCode:   fc,
		SlaveID:        actuator.SlaveID,
		Address:        actuator.RegisterAddress,
		PayloadWritten: value,
		LatencyMs:      latency,
		ExecutedAt:     time.Now().UTC(),
	}

	if writeErr != nil {
		logEntry.Status = model.ActuationStatusFailed
		logEntry.ErrorMessage = writeErr.Error()
		if txErr, ok := writeErr.(*modbus.ModbusTransactionError); ok {
			logEntry.RawRequestHex = hex.EncodeToString(txErr.RequestADU)
			logEntry.RawResponseHex = hex.EncodeToString(txErr.ResponseADU)
		}
		_ = s.actuatorRepo.CreateLog(logEntry)
		_ = s.actuatorRepo.UpdateActuatorState(actuatorID, actuator.CurrentState, actuator.AutoOffDueAt, triggeredBy, string(model.ActuationStatusFailed))
		return logEntry, writeErr
	}

	logEntry.Status = model.ActuationStatusSuccess
	logEntry.RawRequestHex = hex.EncodeToString(resp.RequestADU)
	logEntry.RawResponseHex = hex.EncodeToString(resp.ResponseADU)
	_ = s.actuatorRepo.CreateLog(logEntry)

	// State update
	var newState model.ActuatorState
	var autoOffDue *time.Time

	if action == model.ActionTurnOn {
		newState = model.ActuatorStateOn
		if actuator.SafetyWatchdogActive && actuator.SafetyTimeoutSeconds > 0 {
			due := time.Now().UTC().Add(time.Duration(actuator.SafetyTimeoutSeconds) * time.Second)
			autoOffDue = &due
		}
	} else {
		newState = model.ActuatorStateOff
		autoOffDue = nil
	}

	_ = s.actuatorRepo.UpdateActuatorState(actuatorID, newState, autoOffDue, triggeredBy, string(model.ActuationStatusSuccess))

	// Broadcast WebSocket update for real-time SCADA operator dashboards
	if s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "ACTUATOR_STATE_CHANGED",
			Timestamp: time.Now().UTC(),
			Data: map[string]interface{}{
				"actuator_id":     actuatorID,
				"device_id":       actuator.DeviceID,
				"actuator_code":   actuator.ActuatorCode,
				"actuator_name":   actuator.ActuatorName,
				"state":           string(newState),
				"auto_off_due_at": autoOffDue,
				"action":          string(action),
				"status":          string(model.ActuationStatusSuccess),
				"triggered_by":    triggeredBy,
			},
		})
	}

	logger.Info("Actuator %s (%s) actuated to %s by %s (latency: %dms)",
		actuator.ActuatorName, actuator.ActuatorCode, action, triggeredBy, latency)

	return logEntry, nil
}

// -------------------------------------------------------------
// Schedule CRUD
// -------------------------------------------------------------

func (s *ActuationService) CreateSchedule(schedule *model.ActuationSchedule) error {
	if schedule.ActuatorID == 0 {
		return fmt.Errorf("actuator_id is required")
	}
	if schedule.ScheduleName == "" {
		return fmt.Errorf("schedule_name is required")
	}
	if schedule.IntervalSeconds <= 0 {
		schedule.IntervalSeconds = 7200 // 2h default
	}
	if schedule.DurationSeconds <= 0 {
		schedule.DurationSeconds = 300 // 5m default
	}

	// Compute initial NextRunAt
	now := time.Now().UTC()
	next := now.Add(time.Duration(schedule.IntervalSeconds) * time.Second)
	schedule.NextRunAt = &next

	return s.actuatorRepo.CreateSchedule(schedule)
}

func (s *ActuationService) UpdateSchedule(schedule *model.ActuationSchedule) error {
	existing, err := s.actuatorRepo.GetScheduleByID(schedule.ID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	existing.ScheduleName = schedule.ScheduleName
	existing.ScheduleType = schedule.ScheduleType
	existing.IntervalSeconds = schedule.IntervalSeconds
	existing.DurationSeconds = schedule.DurationSeconds
	existing.StartTimeStr = schedule.StartTimeStr
	existing.EndTimeStr = schedule.EndTimeStr
	existing.DaysOfWeek = schedule.DaysOfWeek
	existing.Enabled = schedule.Enabled

	return s.actuatorRepo.UpdateSchedule(existing)
}

func (s *ActuationService) DeleteSchedule(id uint) error {
	return s.actuatorRepo.DeleteSchedule(id)
}

func (s *ActuationService) ListSchedulesByActuator(actuatorID uint) ([]model.ActuationSchedule, error) {
	return s.actuatorRepo.ListSchedulesByActuator(actuatorID)
}

func (s *ActuationService) GetSchedule(id uint) (*model.ActuationSchedule, error) {
	return s.actuatorRepo.GetScheduleByID(id)
}

func (s *ActuationService) ListLogs(actuatorID, deviceID *uint, limit int) ([]model.ActuationLog, error) {
	return s.actuatorRepo.ListLogs(actuatorID, deviceID, limit)
}
