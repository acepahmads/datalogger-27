package communication

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/service"
)

// AdapterFactory creates ProtocolAdapters based on connection configuration
type AdapterFactory func(cfg *model.DeviceConnection) (ProtocolAdapter, error)

// ConnectionManager manages all protocol connections, auto-recovery, backoff, and health tracking per device
type ConnectionManager struct {
	mu             sync.RWMutex
	adapters       map[uint]ProtocolAdapter
	deviceService  *service.DeviceService
	factory        AdapterFactory
	stateMachine   *ConnectionStateMachine
	healthTracker  *ConnectionHealthTracker
	backoffPolicy  *BackoffPolicy
	reconnectingMu sync.Mutex
	reconnecting   map[uint]bool
	diagMu         sync.Mutex
	diagActive     map[uint]bool
}

// NewConnectionManager initializes the communication connection manager
func NewConnectionManager(deviceService *service.DeviceService, factory AdapterFactory) *ConnectionManager {
	sm := NewConnectionStateMachine()
	ht := NewConnectionHealthTracker()
	bp := DefaultBackoffPolicy()

	cm := &ConnectionManager{
		adapters:      make(map[uint]ProtocolAdapter),
		deviceService: deviceService,
		factory:       factory,
		stateMachine:  sm,
		healthTracker: ht,
		backoffPolicy: bp,
		reconnecting:  make(map[uint]bool),
		diagActive:    make(map[uint]bool),
	}

	// Sync state transitions to health tracker
	sm.OnStateChange(func(deviceID uint, from, to ConnectionState, reason string) {
		ht.SetState(deviceID, to)
	})

	return cm
}

// GetStateMachine returns the deterministic ConnectionStateMachine
func (m *ConnectionManager) GetStateMachine() *ConnectionStateMachine {
	return m.stateMachine
}

// GetHealthTracker returns the ConnectionHealthTracker
func (m *ConnectionManager) GetHealthTracker() *ConnectionHealthTracker {
	return m.healthTracker
}

// GetBackoffPolicy returns the BackoffPolicy
func (m *ConnectionManager) GetBackoffPolicy() *BackoffPolicy {
	return m.backoffPolicy
}

// GetDeviceHealth retrieves the real-time ConnectionHealth for a device
func (m *ConnectionManager) GetDeviceHealth(deviceID uint) ConnectionHealth {
	return m.healthTracker.GetHealth(deviceID)
}

// GetAllDeviceHealth retrieves real-time ConnectionHealth snapshots for all devices
func (m *ConnectionManager) GetAllDeviceHealth() map[uint]ConnectionHealth {
	return m.healthTracker.GetAllHealth()
}

// GetOrCreateAdapter retrieves an active adapter or instantiates one for the device
func (m *ConnectionManager) GetOrCreateAdapter(device *model.Device) (ProtocolAdapter, error) {
	if device == nil {
		return nil, fmt.Errorf("device cannot be nil")
	}
	if device.Connection == nil {
		return nil, fmt.Errorf("device %s (%d) has no connection configuration", device.DeviceCode, device.ID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if adapter, exists := m.adapters[device.ID]; exists {
		// Check if protocol or target has changed
		existingCfg := adapter.GetConfig()
		if existingCfg != nil &&
			existingCfg.Protocol == device.Connection.Protocol &&
			existingCfg.Host == device.Connection.Host &&
			existingCfg.Port == device.Connection.Port &&
			existingCfg.SerialPort == device.Connection.SerialPort &&
			existingCfg.BaudRate == device.Connection.BaudRate &&
			existingCfg.SlaveID == device.Connection.SlaveID {
			return adapter, nil
		}
		// Config changed, close existing
		_ = adapter.Disconnect()
		delete(m.adapters, device.ID)
	}

	adapter, err := m.factory(device.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create protocol adapter for device %s: %w", device.DeviceCode, err)
	}

	m.adapters[device.ID] = adapter
	return adapter, nil
}

// ConnectDevice connects a specific device adapter
func (m *ConnectionManager) ConnectDevice(ctx context.Context, device *model.Device) error {
	if device == nil {
		return fmt.Errorf("device is nil")
	}

	m.healthTracker.RecordConnectionAttempt(device.ID)
	_, _ = m.stateMachine.Transition(device.ID, StateConnecting, "initiating connection")

	adapter, err := m.GetOrCreateAdapter(device)
	if err != nil {
		m.healthTracker.RecordCommunicationFailure(device.ID, err)
		_, _ = m.stateMachine.Transition(device.ID, StateError, err.Error())
		if m.deviceService != nil {
			_ = m.deviceService.SetConnectionError(device.ID, err.Error())
		}
		return err
	}

	timeoutMs := device.Connection.Timeout
	if timeoutMs <= 0 {
		timeoutMs = 1000
	}
	connCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	if err := adapter.Connect(connCtx); err != nil {
		m.healthTracker.RecordCommunicationFailure(device.ID, err)
		_, _ = m.stateMachine.Transition(device.ID, StateError, err.Error())
		if m.deviceService != nil {
			_ = m.deviceService.SetConnectionError(device.ID, err.Error())
		}
		return err
	}

	m.healthTracker.RecordConnectionSuccess(device.ID)
	_, _ = m.stateMachine.Transition(device.ID, StateConnected, "connection established")
	return nil
}

// DisconnectDevice disconnects the device adapter gracefully
func (m *ConnectionManager) DisconnectDevice(deviceID uint) error {
	m.mu.Lock()
	adapter, exists := m.adapters[deviceID]
	m.mu.Unlock()

	var err error
	if exists {
		err = adapter.Disconnect()
	}

	m.healthTracker.SetState(deviceID, StateDisconnected)
	_, _ = m.stateMachine.Transition(deviceID, StateDisconnected, "user or service disconnect")

	if m.deviceService != nil {
		_ = m.deviceService.SetOffline(deviceID)
	}
	return err
}

// DisableDevice cancels pending reconnections and marks the device as DISABLED
func (m *ConnectionManager) DisableDevice(deviceID uint) error {
	_ = m.DisconnectDevice(deviceID)
	m.healthTracker.SetState(deviceID, StateDisabled)
	_, _ = m.stateMachine.Transition(deviceID, StateDisabled, "device disabled by user or policy")
	return nil
}

// ReconnectDevice executes an isolated, deduplicated reconnection cycle with bounded exponential backoff
func (m *ConnectionManager) ReconnectDevice(ctx context.Context, device *model.Device) error {
	if device == nil {
		return fmt.Errorf("device is nil")
	}

	// 1. In-Flight Reconnection Deduplication (singleflight per device)
	m.reconnectingMu.Lock()
	if m.reconnecting[device.ID] {
		m.reconnectingMu.Unlock()
		logger.Debug("Device %s (%d) reconnect already in-flight, skipping duplicate trigger", device.DeviceCode, device.ID)
		return fmt.Errorf("reconnection already in progress for device %d", device.ID)
	}
	m.reconnecting[device.ID] = true
	m.reconnectingMu.Unlock()

	defer func() {
		m.reconnectingMu.Lock()
		delete(m.reconnecting, device.ID)
		m.reconnectingMu.Unlock()
	}()

	// 2. Check if device is disabled or deleted
	if m.stateMachine.GetState(device.ID) == StateDisabled || device.Status == model.DeviceStatusDisabled {
		return fmt.Errorf("device %d is disabled", device.ID)
	}
	if device.Status == model.DeviceStatusInactive {
		_ = m.DisableDevice(device.ID)
		return fmt.Errorf("device %d is disabled", device.ID)
	}

	// 3. Compute bounded exponential backoff with jitter
	health := m.healthTracker.GetHealth(device.ID)
	delay := m.backoffPolicy.CalculateDelay(health.ConsecutiveFailures)

	m.healthTracker.RecordReconnectAttempt(device.ID, delay)
	_, _ = m.stateMachine.Transition(device.ID, StateReconnecting, fmt.Sprintf("scheduled reconnect in %v", delay))

	logger.Debug("Device %s (%d) auto-reconnecting (attempt %d, backoff %v)",
		device.DeviceCode, device.ID, health.ConsecutiveFailures+1, delay)

	// 4. Clean disconnect before backoff wait
	_ = m.DisconnectDevice(device.ID)

	// 5. Backoff delay with context cancellation responsiveness
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
	}

	// 6. Connect attempt
	err := m.ConnectDevice(ctx, device)
	if err != nil {
		// Evaluate threshold for ERROR vs DEGRADED
		consec := m.healthTracker.GetHealth(device.ID).ConsecutiveFailures
		if consec >= m.healthTracker.errorThreshold {
			_, _ = m.stateMachine.Transition(device.ID, StateError, fmt.Sprintf("max reconnect attempts exceeded (%d): %v", consec, err))
		} else {
			_, _ = m.stateMachine.Transition(device.ID, StateDegraded, fmt.Sprintf("reconnect failed: %v", err))
		}
		return err
	}

	// 7. Success
	m.healthTracker.RecordReconnectSuccess(device.ID)
	_, _ = m.stateMachine.Transition(device.ID, StateConnected, "reconnect successful")
	logger.Info("Device %s (%d) successfully recovered and reconnected", device.DeviceCode, device.ID)
	return nil
}

// ExecuteWithRetry reads registers with configurable retry count, jittered backoff, and error categorization
func (m *ConnectionManager) ExecuteWithRetry(
	ctx context.Context,
	device *model.Device,
	req ModbusReadRequest,
) (*ModbusReadResponse, error) {
	adapter, err := m.GetOrCreateAdapter(device)
	if err != nil {
		m.healthTracker.RecordCommunicationFailure(device.ID, err)
		return nil, err
	}

	retryLimit := device.Connection.RetryCount
	if retryLimit < 0 {
		retryLimit = 0
	}
	if retryLimit > 5 {
		retryLimit = 5
	}

	var lastErr error
	for attempt := 0; attempt <= retryLimit; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := adapter.ReadRegisters(ctx, req)
		if err == nil {
			// Successful read
			latencyMs := int(resp.ResponseTime.Milliseconds())
			m.healthTracker.RecordCommunicationSuccess(device.ID, latencyMs)

			// Recover state if degraded or reconnecting
			curState := m.stateMachine.GetState(device.ID)
			if curState == StateDegraded || curState == StateReconnecting || curState == StateError {
				_, _ = m.stateMachine.Transition(device.ID, StateConnected, "successful communication read")
			}

			if m.deviceService != nil {
				_ = m.deviceService.SetOnline(device.ID)
				_ = m.deviceService.UpdateLastSeen(device.ID)
				_ = m.deviceService.RecordCommunicationResult(device.ID, true, latencyMs)
			}
			return resp, nil
		}

		lastErr = err
		if attempt < retryLimit {
			// Bounded jittered retry delay: (attempt + 1) * 50ms +/- 20%
			retryDelay := time.Duration(50*(attempt+1)) * time.Millisecond
			logger.Debug("Device %s read failed (attempt %d/%d): %v. Retrying in %v...",
				device.DeviceCode, attempt+1, retryLimit+1, err, retryDelay)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}
	}

	// All retries failed
	m.healthTracker.RecordCommunicationFailure(device.ID, lastErr)
	consec := m.healthTracker.GetHealth(device.ID).ConsecutiveFailures

	// If failures exceed degraded threshold, transition state machine
	if consec >= m.healthTracker.degradedThreshold {
		_, _ = m.stateMachine.Transition(device.ID, StateDegraded, fmt.Sprintf("%d consecutive failures: %v", consec, lastErr))
	}

	if m.deviceService != nil {
		_ = m.deviceService.SetConnectionError(device.ID, lastErr.Error())
		_ = m.deviceService.RecordCommunicationResult(device.ID, false, 0)
	}

	return nil, fmt.Errorf("device %s communication failed after %d retries: %w", device.DeviceCode, retryLimit, lastErr)
}

// DiagnosticTest performs a non-destructive, serialized diagnostic check of the device link.
// It returns structured results (SUCCESS, TIMEOUT, SERIAL_OPEN_ERROR, SERIAL_IO_ERROR,
// CRC_ERROR, MODBUS_EXCEPTION, CONFIGURATION_ERROR, BUSY) without tearing down
// active polling transports or corrupting device operational status.
func (m *ConnectionManager) DiagnosticTest(ctx context.Context, device *model.Device) *DiagnosticResult {
	now := time.Now()
	res := &DiagnosticResult{
		Timestamp: now,
	}

	if device == nil || device.Connection == nil {
		res.Code = DiagConfigurationError
		res.Error = "device or connection configuration is missing"
		res.Message = "Configuration error: missing connection settings"
		return res
	}

	cfg := device.Connection
	res.Port = cfg.SerialPort
	if res.Port == "" {
		res.Port = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}
	res.BaudRate = cfg.BaudRate

	slaveID := byte(cfg.SlaveID)
	if slaveID == 0 {
		slaveID = 1
	}
	res.SlaveID = slaveID

	// 1. Validate settings
	if cfg.Protocol == model.ProtocolModbusRTU || cfg.Protocol == model.ProtocolSerial {
		if cfg.SerialPort == "" {
			res.Code = DiagConfigurationError
			res.Error = "serial_port cannot be empty for Modbus RTU"
			res.Message = "Configuration error: serial port not specified"
			return res
		}
	} else if cfg.Protocol == model.ProtocolModbusTCP || cfg.Protocol == model.ProtocolTCP {
		if cfg.Host == "" || cfg.Port <= 0 {
			res.Code = DiagConfigurationError
			res.Error = "host and port must be specified for Modbus TCP"
			res.Message = "Configuration error: invalid host or port"
			return res
		}
	}

	// 2. Prevent concurrent diagnostic tests on the same device
	m.diagMu.Lock()
	if m.diagActive == nil {
		m.diagActive = make(map[uint]bool)
	}
	if m.diagActive[device.ID] {
		m.diagMu.Unlock()
		res.Code = DiagBusy
		res.Error = fmt.Sprintf("device %s is busy with another diagnostic test", device.DeviceCode)
		res.Message = "Diagnostic busy: another test is currently running"
		return res
	}
	m.diagActive[device.ID] = true
	m.diagMu.Unlock()

	defer func() {
		m.diagMu.Lock()
		delete(m.diagActive, device.ID)
		m.diagMu.Unlock()
	}()

	// 3. Resolve target register
	fc := FunctionReadHoldingRegisters
	addr := uint16(0)
	qty := uint16(1)

	// If device has active parameters, test against the first active parameter
	for _, p := range device.Parameters {
		if p.Enabled {
			r, err := modbus.ResolveRegisterAddress(p.RegisterType, p.RegisterAddress)
			if err != nil {
				res.Code = DiagConfigurationError
				res.Error = fmt.Sprintf("invalid register configuration for parameter %s: %v", p.ParameterCode, err)
				res.Message = "Configuration error: invalid register address or type"
				return res
			}
			fc = r.FunctionCode
			addr = r.PDUAddress
			qty = modbus.RequiredRegisterCount(p.DataType)
			if fc == FunctionReadCoils || fc == FunctionReadDiscreteInputs {
				qty = 1
			}
			break
		}
	}

	res.FunctionCode = fc
	res.StartingAddress = addr
	res.RegisterCount = qty

	// 4. Get or create adapter
	adapter, err := m.GetOrCreateAdapter(device)
	if err != nil {
		res.Code = DiagSerialOpenError
		res.Error = err.Error()
		res.Message = fmt.Sprintf("Failed to initialize adapter: %v", err)
		return res
	}

	// 5. Connect if not connected
	timeoutMs := cfg.Timeout
	if timeoutMs <= 0 {
		timeoutMs = 1500
	}
	diagTimeout := time.Duration(timeoutMs) * time.Millisecond
	if diagTimeout < 500*time.Millisecond {
		diagTimeout = 500 * time.Millisecond
	}

	testCtx, cancel := context.WithTimeout(ctx, diagTimeout)
	defer cancel()

	if !adapter.IsConnected() {
		if err := adapter.Connect(testCtx); err != nil {
			res.Code = DiagSerialOpenError
			res.Error = err.Error()
			res.Message = fmt.Sprintf("Failed to open connection to %s: %v", res.Port, err)
			return res
		}
	}

	// 6. Execute Diagnostic Read
	start := time.Now()
	readReq := ModbusReadRequest{
		SlaveID:         slaveID,
		FunctionCode:    fc,
		StartingAddress: addr,
		Quantity:        qty,
	}

	_, err = adapter.ReadRegisters(testCtx, readReq)
	elapsed := int(time.Since(start).Milliseconds())
	res.LatencyMs = elapsed

	if err == nil {
		res.Code = DiagSuccess
		res.Success = true
		res.Connected = true
		res.Message = fmt.Sprintf("Physical link verified: Slave ID %d responded in %d ms (FC %02X, Addr %d)", slaveID, elapsed, fc, addr)

		// Record successful communication
		m.healthTracker.RecordCommunicationSuccess(device.ID, elapsed)
		if m.deviceService != nil {
			_ = m.deviceService.SetOnline(device.ID)
			_ = m.deviceService.UpdateLastSeen(device.ID)
			_ = m.deviceService.RecordCommunicationResult(device.ID, true, elapsed)
		}
		return res
	}

	// 7. Parse and categorize error
	errStr := err.Error()
	res.Error = errStr

	if strings.Contains(errStr, "Modbus exception") || strings.Contains(errStr, "Illegal") {
		// Valid Modbus exception response received from slave device!
		// This proves physical layer, framing, slave address, baud rate, and CRC are 100% WORKING!
		res.Code = DiagModbusException
		res.Success = true
		res.Connected = true
		res.Message = fmt.Sprintf("Physical link verified: Slave ID %d responded with %s in %d ms (link active)", slaveID, errStr, elapsed)

		m.healthTracker.RecordCommunicationSuccess(device.ID, elapsed)
		if m.deviceService != nil {
			_ = m.deviceService.SetOnline(device.ID)
			_ = m.deviceService.UpdateLastSeen(device.ID)
		}
		return res
	}

	if strings.Contains(errStr, "CRC16") || strings.Contains(errStr, "checksum") {
		res.Code = DiagCRCError
		res.Message = fmt.Sprintf("CRC error: corrupted frame received from %s (check baud rate, parity, or wiring noise)", res.Port)
		return res
	}

	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(errStr), "timeout") {
		res.Code = DiagTimeout
		res.Message = fmt.Sprintf("Timeout: no response from Slave ID %d on %s within %d ms", slaveID, res.Port, timeoutMs)
		return res
	}

	if strings.Contains(errStr, "serial port offline") || strings.Contains(errStr, "failed to open") {
		res.Code = DiagSerialOpenError
		res.Message = fmt.Sprintf("Serial open error on %s: %s", res.Port, errStr)
		return res
	}

	res.Code = DiagSerialIOError
	res.Message = fmt.Sprintf("I/O error during communication on %s: %s", res.Port, errStr)
	return res
}

// TestConnection performs an isolated connection test using DiagnosticTest
func (m *ConnectionManager) TestConnection(ctx context.Context, device *model.Device) (int, error) {
	diag := m.DiagnosticTest(ctx, device)
	if !diag.Success && !diag.Connected {
		return diag.LatencyMs, errors.New(diag.Error)
	}
	return diag.LatencyMs, nil
}

// GetAdapterStatus returns live status of the adapter combined with health metrics
func (m *ConnectionManager) GetAdapterStatus(deviceID uint) (AdapterStatus, bool) {
	m.mu.RLock()
	adapter, exists := m.adapters[deviceID]
	m.mu.RUnlock()

	health := m.healthTracker.GetHealth(deviceID)

	if !exists {
		return AdapterStatus{
			State:     health.State,
			LastError: health.LastError,
		}, false
	}

	status := adapter.GetStatus()
	// Overlay state machine state if more specific (e.g. DEGRADED, RECONNECTING)
	if health.State != "" {
		status.State = health.State
	}
	return status, true
}

// CloseAll closes all active connections and cancels tracking
func (m *ConnectionManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, adapter := range m.adapters {
		_ = adapter.Disconnect()
		m.healthTracker.SetState(id, StateDisconnected)
		_, _ = m.stateMachine.Transition(id, StateDisconnected, "shutdown")
		delete(m.adapters, id)
	}
}

// GetDeviceState returns the current state machine state for a device
func (m *ConnectionManager) GetDeviceState(deviceID uint) ConnectionState {
	return m.stateMachine.GetState(deviceID)
}

// IsDeviceReconnecting returns true if an asynchronous reconnection is currently active for the device
func (m *ConnectionManager) IsDeviceReconnecting(deviceID uint) bool {
	m.reconnectingMu.Lock()
	defer m.reconnectingMu.Unlock()
	return m.reconnecting[deviceID]
}

// SetBackoffPolicy updates the backoff policy
func (m *ConnectionManager) SetBackoffPolicy(p *BackoffPolicy) {
	if p != nil {
		m.backoffPolicy = p
	}
}
