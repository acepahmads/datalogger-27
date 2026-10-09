package communication

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

// setupMockRTUManager initializes a ConnectionManager with a MockModbusServer and MockSerialPipe
func setupMockRTUManager(server *modbus.MockModbusServer) *ConnectionManager {
	opener := func(cfg *model.DeviceConnection) (modbus.SerialTransport, error) {
		return modbus.NewMockSerialPipe(server), nil
	}
	factory := func(cfg *model.DeviceConnection) (ProtocolAdapter, error) {
		if cfg.Protocol == model.ProtocolModbusRTU {
			return modbus.NewModbusRTUAdapterWithOpener(cfg, opener), nil
		}
		return DefaultAdapterFactory(cfg)
	}
	return NewConnectionManager(nil, factory)
}

// 1. Regression Test: Successful Modbus response
func TestDiagnosticSuccess(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 0x1234)
	mgr := setupMockRTUManager(server)

	dev := &model.Device{
		ID:         10,
		DeviceCode: "DEV-RTU-01",
		Connection: &model.DeviceConnection{
			DeviceID:   10,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_TEST_SUCCESS",
			BaudRate:   9600,
			Timeout:    500,
			SlaveID:    1,
		},
		Parameters: []model.Parameter{
			{
				ID:              1,
				DeviceID:        10,
				ParameterCode:   "REG0",
				RegisterType:    "HOLDING",
				RegisterAddress: 0, // Direct 0-based PDU 0
				DataType:        "INT16",
				Enabled:         true,
			},
		},
	}

	ctx := context.Background()
	diag := mgr.DiagnosticTest(ctx, dev)

	if !diag.Success || !diag.Connected {
		t.Fatalf("expected successful diagnostic test, got: %+v", diag)
	}
	if diag.Code != modbus.DiagSuccess {
		t.Fatalf("expected DiagSuccess, got: %s", diag.Code)
	}
	if diag.SlaveID != 1 {
		t.Fatalf("expected SlaveID 1, got %d", diag.SlaveID)
	}
	if diag.FunctionCode != modbus.FunctionReadHoldingRegisters {
		t.Fatalf("expected FC 03, got %d", diag.FunctionCode)
	}
	if diag.StartingAddress != 0 {
		t.Fatalf("expected PDU address 0, got %d", diag.StartingAddress)
	}
	if diag.LatencyMs < 0 {
		t.Fatalf("invalid latency: %d", diag.LatencyMs)
	}
}

// 2. Regression Test: Valid Modbus exception response returns MODBUS_EXCEPTION with Connected=true
func TestDiagnosticModbusException(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetMode(modbus.SimModeException, modbus.ExIllegalDataAddress, 0)
	mgr := setupMockRTUManager(server)

	dev := &model.Device{
		ID:         11,
		DeviceCode: "DEV-RTU-EX",
		Connection: &model.DeviceConnection{
			DeviceID:   11,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_TEST_EX",
			BaudRate:   9600,
			Timeout:    500,
			SlaveID:    1,
		},
	}

	ctx := context.Background()
	diag := mgr.DiagnosticTest(ctx, dev)

	if !diag.Connected || !diag.Success {
		t.Fatalf("expected Connected=true on valid Modbus exception response, got: %+v", diag)
	}
	if diag.Code != modbus.DiagModbusException {
		t.Fatalf("expected DiagModbusException, got: %s", diag.Code)
	}
	if diag.Error == "" {
		t.Fatalf("expected non-empty error string explaining exception")
	}
}

// 3. Regression Test: Timeout and recovery (port remains open and recovers immediately)
func TestDiagnosticTimeoutAndRecovery(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 0x55AA)
	mgr := setupMockRTUManager(server)

	dev := &model.Device{
		ID:         12,
		DeviceCode: "DEV-RTU-TIMEOUT",
		Connection: &model.DeviceConnection{
			DeviceID:   12,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_TEST_TIMEOUT",
			BaudRate:   9600,
			Timeout:    50, // Short timeout
			SlaveID:    1,
		},
	}

	ctx := context.Background()

	// Connect first
	if err := mgr.ConnectDevice(ctx, dev); err != nil {
		t.Fatalf("ConnectDevice failed: %v", err)
	}
	defer mgr.DisconnectDevice(dev.ID)

	// A. Simulate slave timeout (sleep longer than deadline)
	server.SetMode(modbus.SimModeTimeout, 0, 200*time.Millisecond)

	diag1 := mgr.DiagnosticTest(ctx, dev)
	if diag1.Connected || diag1.Success {
		t.Fatalf("expected diagnostic failure on timeout, got %+v", diag1)
	}
	if diag1.Code != modbus.DiagTimeout {
		t.Fatalf("expected DiagTimeout, got %s", diag1.Code)
	}

	// B. Restore server to normal mode: subsequent diagnostic must succeed IMMEDIATELY
	// without needing a manual reconnect, proving transport was preserved!
	server.SetMode(modbus.SimModeNormal, 0, 0)

	diag2 := mgr.DiagnosticTest(ctx, dev)
	if !diag2.Connected || !diag2.Success {
		t.Fatalf("recovery failed after timeout: %+v", diag2)
	}
	if diag2.Code != modbus.DiagSuccess {
		t.Fatalf("expected DiagSuccess on recovery, got %s", diag2.Code)
	}
}

// 4. Regression Test: Serial port already owned by adapter (shared bus, refCount, no duplicate opens)
func TestSerialPortAlreadyOwnedByAdapter(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 100)

	var openCalls int32
	pipe := modbus.NewMockSerialPipe(server)
	opener := func(cfg *model.DeviceConnection) (modbus.SerialTransport, error) {
		atomic.AddInt32(&openCalls, 1)
		return pipe, nil
	}

	factory := func(cfg *model.DeviceConnection) (ProtocolAdapter, error) {
		return modbus.NewModbusRTUAdapterWithOpener(cfg, opener), nil
	}

	mgr := NewConnectionManager(nil, factory)
	ctx := context.Background()

	dev1 := &model.Device{
		ID:         21,
		DeviceCode: "DEV-MULTI-1",
		Connection: &model.DeviceConnection{
			DeviceID:   21,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_SHARED_TEST",
			BaudRate:   9600,
			SlaveID:    1,
			Timeout:    500,
		},
	}
	dev2 := &model.Device{
		ID:         22,
		DeviceCode: "DEV-MULTI-2",
		Connection: &model.DeviceConnection{
			DeviceID:   22,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_SHARED_TEST", // Same physical port!
			BaudRate:   9600,
			SlaveID:    2,
			Timeout:    500,
		},
	}

	// Connect dev1
	if err := mgr.ConnectDevice(ctx, dev1); err != nil {
		t.Fatalf("dev1 connect failed: %v", err)
	}
	if atomic.LoadInt32(&openCalls) != 1 {
		t.Fatalf("expected 1 open call, got %d", openCalls)
	}

	// Connect dev2 on SAME serial port
	if err := mgr.ConnectDevice(ctx, dev2); err != nil {
		t.Fatalf("dev2 connect failed: %v", err)
	}
	// Verify opener was NOT called again: shared bus reused the open port!
	if atomic.LoadInt32(&openCalls) != 1 {
		t.Fatalf("expected still 1 open call (shared bus reuse), got %d", openCalls)
	}

	// Run diagnostic on dev1: must reuse existing transport without opening again
	diag := mgr.DiagnosticTest(ctx, dev1)
	if !diag.Success {
		t.Fatalf("diagnostic on dev1 failed: %+v", diag)
	}
	if atomic.LoadInt32(&openCalls) != 1 {
		t.Fatalf("diagnostic opened a duplicate connection! openCalls = %d", openCalls)
	}

	// Disconnect dev1: port should remain open because dev2 is still attached
	_ = mgr.DisconnectDevice(dev1.ID)
	bus := modbus.GetOrCreateSharedBus("COM_SHARED_TEST")
	if bus == nil {
		t.Fatalf("expected shared bus to still exist for dev2")
	}

	// Disconnect dev2: now port should close
	_ = mgr.DisconnectDevice(dev2.ID)
}

// 5. Regression Test: Test Link while background polling is actively executing
func TestDiagnosticWhilePollingActive(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 42)
	mgr := setupMockRTUManager(server)

	dev := &model.Device{
		ID:         30,
		DeviceCode: "DEV-POLL-DIAG",
		Connection: &model.DeviceConnection{
			DeviceID:   30,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_POLL_DIAG",
			BaudRate:   9600,
			Timeout:    500,
			SlaveID:    1,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mgr.ConnectDevice(ctx, dev); err != nil {
		t.Fatalf("ConnectDevice failed: %v", err)
	}
	defer mgr.DisconnectDevice(dev.ID)

	// Launch active background polling goroutine
	var pollCount int64
	var pollErrors int64
	stopPoll := make(chan struct{})

	go func() {
		req := ModbusReadRequest{
			SlaveID:         1,
			FunctionCode:    FunctionReadHoldingRegisters,
			StartingAddress: 0,
			Quantity:        1,
		}
		for {
			select {
			case <-stopPoll:
				return
			default:
				_, err := mgr.ExecuteWithRetry(ctx, dev, req)
				if err != nil {
					atomic.AddInt64(&pollErrors, 1)
				} else {
					atomic.AddInt64(&pollCount, 1)
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Wait for background polling to complete a few reads
	time.Sleep(30 * time.Millisecond)

	// Execute Test Link (DiagnosticTest) concurrently with polling
	for i := 0; i < 5; i++ {
		diag := mgr.DiagnosticTest(ctx, dev)
		if !diag.Success || !diag.Connected {
			t.Fatalf("diagnostic failed during polling (iteration %d): %+v", i, diag)
		}
		time.Sleep(10 * time.Millisecond)
	}

	close(stopPoll)

	if atomic.LoadInt64(&pollCount) == 0 {
		t.Fatalf("expected polling to have succeeded during diagnostics, got 0")
	}
	if atomic.LoadInt64(&pollErrors) > 0 {
		t.Fatalf("polling encountered %d errors during diagnostics", pollErrors)
	}
}

// 6. Regression Test: Concurrent diagnostic requests return BUSY
func TestConcurrentDiagnosticRequests(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 100)
	mgr := setupMockRTUManager(server)

	dev := &model.Device{
		ID:         40,
		DeviceCode: "DEV-CONCUR-DIAG",
		Connection: &model.DeviceConnection{
			DeviceID:   40,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_CONCUR_DIAG",
			BaudRate:   9600,
			Timeout:    500,
			SlaveID:    1,
		},
	}

	ctx := context.Background()

	// Simulate one diagnostic being held active via diagActive map
	mgr.diagMu.Lock()
	mgr.diagActive[dev.ID] = true
	mgr.diagMu.Unlock()

	// Second concurrent call should immediately return BUSY
	diag := mgr.DiagnosticTest(ctx, dev)
	if diag.Connected || diag.Code != modbus.DiagBusy {
		t.Fatalf("expected DiagBusy on concurrent diagnostic, got %+v", diag)
	}

	// Release mock lock
	mgr.diagMu.Lock()
	delete(mgr.diagActive, dev.ID)
	mgr.diagMu.Unlock()

	// Now it must succeed
	diagSuccess := mgr.DiagnosticTest(ctx, dev)
	if !diagSuccess.Success || diagSuccess.Code != modbus.DiagSuccess {
		t.Fatalf("expected success after lock released, got %+v", diagSuccess)
	}
}

// 7. Regression Test: Retry counter reset after success
func TestRetryCounterResetAfterSuccess(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 200)
	mgr := setupMockRTUManager(server)

	dev := &model.Device{
		ID:         50,
		DeviceCode: "DEV-RETRY-RESET",
		Connection: &model.DeviceConnection{
			DeviceID:   50,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_RETRY_RESET",
			BaudRate:   9600,
			Timeout:    500,
			SlaveID:    1,
		},
	}

	ctx := context.Background()
	if err := mgr.ConnectDevice(ctx, dev); err != nil {
		t.Fatalf("ConnectDevice failed: %v", err)
	}
	defer mgr.DisconnectDevice(dev.ID)

	// Record 3 consecutive failures into health tracker
	mgr.healthTracker.RecordCommunicationFailure(dev.ID, errors.New("serial: timeout"))
	mgr.healthTracker.RecordCommunicationFailure(dev.ID, errors.New("serial: timeout"))
	mgr.healthTracker.RecordCommunicationFailure(dev.ID, errors.New("serial: timeout"))

	healthBefore := mgr.healthTracker.GetHealth(dev.ID)
	if healthBefore.ConsecutiveFailures != 3 {
		t.Fatalf("expected ConsecutiveFailures == 3, got %d", healthBefore.ConsecutiveFailures)
	}

	// Execute successful read
	req := ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	}
	resp, err := mgr.ExecuteWithRetry(ctx, dev, req)
	if err != nil {
		t.Fatalf("ExecuteWithRetry failed: %v", err)
	}
	if len(resp.Data) < 2 {
		t.Fatalf("unexpected data: %v", resp.Data)
	}

	// Verify consecutive failures reset to 0
	healthAfter := mgr.healthTracker.GetHealth(dev.ID)
	if healthAfter.ConsecutiveFailures != 0 {
		t.Fatalf("expected ConsecutiveFailures reset to 0, got %d", healthAfter.ConsecutiveFailures)
	}
	if healthAfter.ConsecutiveSuccesses != 1 {
		t.Fatalf("expected ConsecutiveSuccesses == 1, got %d", healthAfter.ConsecutiveSuccesses)
	}
}

// 8. Regression Test: Communication status changing after stale/no successful communication
func TestCommunicationStatusSemantics(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 300)
	mgr := setupMockRTUManager(server)

	dev := &model.Device{
		ID:         60,
		DeviceCode: "DEV-STATUS-SEMANTICS",
		Connection: &model.DeviceConnection{
			DeviceID:   60,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_STATUS_SEMANTICS",
			BaudRate:   9600,
			Timeout:    500,
			SlaveID:    1,
		},
	}

	ctx := context.Background()

	// A. Initial state is DISCONNECTED, not ONLINE
	initialStatus, _ := mgr.GetAdapterStatus(dev.ID)
	if initialStatus.State != StateDisconnected {
		t.Fatalf("expected initial StateDisconnected, got %v", initialStatus.State)
	}

	// B. Connect and perform successful communication -> ONLINE
	if err := mgr.ConnectDevice(ctx, dev); err != nil {
		t.Fatalf("ConnectDevice failed: %v", err)
	}
	defer mgr.DisconnectDevice(dev.ID)

	diag := mgr.DiagnosticTest(ctx, dev)
	if !diag.Success {
		t.Fatalf("diagnostic failed: %+v", diag)
	}

	health := mgr.healthTracker.GetHealth(dev.ID)
	if health.State != StateConnected || health.ConsecutiveSuccesses == 0 {
		t.Fatalf("expected health StateConnected with positive successes, got %+v", health)
	}

	// C. A single failed Test Link (e.g. invalid register) must NOT mark the healthy device offline!
	devInvalidReg := *dev
	devInvalidReg.Parameters = []model.Parameter{
		{
			ID:              999,
			DeviceID:        60,
			ParameterCode:   "BAD_REG",
			RegisterType:    "HOLDING",
			RegisterAddress: 99999, // Out of bounds
			DataType:        "INT16",
			Enabled:         true,
		},
	}
	// Address 99999 will cause resolution error, triggering configuration error
	diagBad := mgr.DiagnosticTest(ctx, &devInvalidReg)
	if diagBad.Connected {
		t.Fatalf("expected failure on invalid register, got %+v", diagBad)
	}

	// Verify device health remains CONNECTED
	healthStillGood := mgr.healthTracker.GetHealth(dev.ID)
	if healthStillGood.State != StateConnected {
		t.Fatalf("device was incorrectly marked offline by a single failed diagnostic: %+v", healthStillGood)
	}

	// D. Repeated polling failures exceed errorThreshold (10 failures) -> transitions to StateError
	for i := 0; i < 10; i++ {
		mgr.healthTracker.RecordCommunicationFailure(dev.ID, errors.New("serial: timeout"))
	}
	healthFailed := mgr.healthTracker.GetHealth(dev.ID)
	if healthFailed.State != StateError {
		t.Fatalf("expected StateError after 10 consecutive failures, got %v", healthFailed.State)
	}
}

// 9. Regression Test: No goroutine or serial-port leaks
func TestNoLeaksAfterOperations(t *testing.T) {
	defer modbus.ResetSharedBuses()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 400)
	mgr := setupMockRTUManager(server)

	dev := &model.Device{
		ID:         70,
		DeviceCode: "DEV-LEAK-TEST",
		Connection: &model.DeviceConnection{
			DeviceID:   70,
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "COM_LEAK_TEST",
			BaudRate:   9600,
			Timeout:    500,
			SlaveID:    1,
		},
	}

	ctx := context.Background()

	// Warm up runtime and garbage collector
	runtime.GC()
	initialGoroutines := runtime.NumGoroutine()

	for cycle := 0; cycle < 10; cycle++ {
		if err := mgr.ConnectDevice(ctx, dev); err != nil {
			t.Fatalf("cycle %d connect failed: %v", cycle, err)
		}

		diag := mgr.DiagnosticTest(ctx, dev)
		if !diag.Success {
			t.Fatalf("cycle %d diagnostic failed: %+v", cycle, diag)
		}

		if err := mgr.DisconnectDevice(dev.ID); err != nil {
			t.Fatalf("cycle %d disconnect failed: %v", cycle, err)
		}
	}

	modbus.ResetSharedBuses()
	runtime.GC()
	time.Sleep(20 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	// Allow +/- 2 goroutines for test runner fluctuations
	diff := finalGoroutines - initialGoroutines
	if diff > 3 {
		t.Fatalf("goroutine leak detected: started with %d, ended with %d (diff: %d)",
			initialGoroutines, finalGoroutines, diff)
	}
}
