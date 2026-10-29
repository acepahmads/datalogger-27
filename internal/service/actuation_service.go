package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
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

	// Compute initial NextRunAt using dynamic schedule type calculation
	now := time.Now()
	next := CalculateNextScheduleRun(schedule, now)
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
	existing.MinuteOfHour = schedule.MinuteOfHour
	existing.StartTimeStr = schedule.StartTimeStr
	existing.EndTimeStr = schedule.EndTimeStr
	existing.CronExpression = schedule.CronExpression
	existing.DaysOfWeek = schedule.DaysOfWeek
	existing.Enabled = schedule.Enabled

	// Recompute NextRunAt based on updated timing configuration
	now := time.Now()
	next := CalculateNextScheduleRun(existing, now)
	existing.NextRunAt = &next

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

// -------------------------------------------------------------
// Advanced Custom Time Determination & Next-Run Engine
// -------------------------------------------------------------

// CalculateNextScheduleRun computes the next trigger timestamp based on schedule type and recurrence rules
func CalculateNextScheduleRun(sched *model.ActuationSchedule, from time.Time) time.Time {
	if sched == nil {
		return from.Add(2 * time.Hour)
	}

	switch sched.ScheduleType {
	case model.ScheduleTypeHourlyMinute:
		return calculateNextHourlyMinute(sched, from)
	case model.ScheduleTypeTimeOfDay:
		return calculateNextTimeOfDay(sched, from)
	case model.ScheduleTypeCron:
		return calculateNextCron(sched, from)
	case model.ScheduleTypeInterval:
		fallthrough
	default:
		intervalSec := sched.IntervalSeconds
		if intervalSec <= 0 {
			intervalSec = 7200
		}
		return from.Add(time.Duration(intervalSec) * time.Second)
	}
}

func parseMinuteList(str string) []int {
	var mins []int
	seen := make(map[int]bool)
	for _, part := range strings.Split(str, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if val, err := strconv.Atoi(part); err == nil && val >= 0 && val <= 59 {
			if !seen[val] {
				seen[val] = true
				mins = append(mins, val)
			}
		}
	}
	if len(mins) == 0 {
		mins = []int{0}
	}
	sort.Ints(mins)
	return mins
}

func calculateNextHourlyMinute(sched *model.ActuationSchedule, from time.Time) time.Time {
	mins := parseMinuteList(sched.MinuteOfHour)
	loc := from.Location()

	// Base truncated to current minute
	base := time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), from.Minute(), 0, 0, loc)

	for hOffset := 0; hOffset < 24*7; hOffset++ {
		hourBase := base.Add(time.Duration(hOffset) * time.Hour)
		for _, m := range mins {
			candidate := time.Date(hourBase.Year(), hourBase.Month(), hourBase.Day(), hourBase.Hour(), m, 0, 0, loc)
			if candidate.After(from) && IsDayStringEnabled(sched.DaysOfWeek, candidate) {
				return candidate
			}
		}
	}
	return from.Add(1 * time.Hour)
}

func parseTimesOfDay(str string) [][2]int {
	var times [][2]int
	for _, part := range strings.Split(str, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		sub := strings.Split(part, ":")
		if len(sub) >= 2 {
			h, e1 := strconv.Atoi(strings.TrimSpace(sub[0]))
			m, e2 := strconv.Atoi(strings.TrimSpace(sub[1]))
			if e1 == nil && e2 == nil && h >= 0 && h <= 23 && m >= 0 && m <= 59 {
				times = append(times, [2]int{h, m})
			}
		}
	}
	if len(times) == 0 {
		times = append(times, [2]int{8, 0})
	}
	return times
}

func calculateNextTimeOfDay(sched *model.ActuationSchedule, from time.Time) time.Time {
	times := parseTimesOfDay(sched.StartTimeStr)
	loc := from.Location()

	for dOffset := 0; dOffset < 8; dOffset++ {
		dayBase := from.AddDate(0, 0, dOffset)
		if !IsDayStringEnabled(sched.DaysOfWeek, dayBase) {
			continue
		}
		var earliestThisDay *time.Time
		for _, t := range times {
			candidate := time.Date(dayBase.Year(), dayBase.Month(), dayBase.Day(), t[0], t[1], 0, 0, loc)
			if candidate.After(from) {
				if earliestThisDay == nil || candidate.Before(*earliestThisDay) {
					c := candidate
					earliestThisDay = &c
				}
			}
		}
		if earliestThisDay != nil {
			return *earliestThisDay
		}
	}
	return from.Add(24 * time.Hour)
}

func calculateNextCron(sched *model.ActuationSchedule, from time.Time) time.Time {
	expr := strings.TrimSpace(sched.CronExpression)
	if expr == "" {
		expr = "0 * * * *"
	}
	loc := from.Location()
	start := from.Truncate(time.Minute).Add(time.Minute)

	for i := 0; i < 7*24*60; i++ {
		curr := start.Add(time.Duration(i) * time.Minute)
		curr = time.Date(curr.Year(), curr.Month(), curr.Day(), curr.Hour(), curr.Minute(), 0, 0, loc)
		if IsDayStringEnabled(sched.DaysOfWeek, curr) && matchCronExpression(expr, curr) {
			return curr
		}
	}
	return from.Add(1 * time.Hour)
}

func matchCronExpression(expr string, t time.Time) bool {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false
	}
	if !matchCronField(fields[0], t.Minute(), 0, 59) {
		return false
	}
	if !matchCronField(fields[1], t.Hour(), 0, 23) {
		return false
	}
	if !matchCronField(fields[2], t.Day(), 1, 31) {
		return false
	}
	if !matchCronField(fields[3], int(t.Month()), 1, 12) {
		return false
	}
	dow := int(t.Weekday()) // 0=Sunday
	if !matchCronField(fields[4], dow, 0, 7) && !(dow == 0 && matchCronField(fields[4], 7, 0, 7)) {
		return false
	}
	return true
}

func matchCronField(field string, val int, minVal, maxVal int) bool {
	field = strings.TrimSpace(field)
	if field == "*" {
		return true
	}
	if strings.Contains(field, ",") {
		for _, part := range strings.Split(field, ",") {
			if matchCronField(part, val, minVal, maxVal) {
				return true
			}
		}
		return false
	}
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err == nil && step > 0 {
			return (val-minVal)%step == 0
		}
		return false
	}
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) == 2 {
			start, e1 := strconv.Atoi(parts[0])
			end, e2 := strconv.Atoi(parts[1])
			if e1 == nil && e2 == nil {
				return val >= start && val <= end
			}
		}
		return false
	}
	num, err := strconv.Atoi(field)
	if err == nil {
		return val == num
	}
	return false
}

// IsDayStringEnabled checks whether a given timestamp matches the ISO 1-7 (Mon-Sun) days_of_week filter
func IsDayStringEnabled(daysOfWeek string, t time.Time) bool {
	if strings.TrimSpace(daysOfWeek) == "" || daysOfWeek == "*" {
		return true
	}
	isoWeekday := int(t.Weekday())
	if isoWeekday == 0 {
		isoWeekday = 7
	}
	dayStr := strconv.Itoa(isoWeekday)
	tokens := strings.Split(daysOfWeek, ",")
	for _, tok := range tokens {
		if strings.TrimSpace(tok) == dayStr {
			return true
		}
	}
	return false
}
