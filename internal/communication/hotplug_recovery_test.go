package communication_test

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"datalogger/internal/communication"
	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

// ControllableMockTransport implements modbus.SerialTransport with physical disconnect simulation
type ControllableMockTransport struct {
	mu         sync.Mutex
	server     *modbus.MockModbusServer
	pipe       *modbus.MockSerialPipe
	portName   string
	unplugged  bool
	ioErr      error
	closeCount int
}

func newControllableTransport(portName string, server *modbus.MockModbusServer) *ControllableMockTransport {
	return &ControllableMockTransport{
		portName: portName,
		server:   server,
		pipe:     modbus.NewMockSerialPipe(server),
	}
}

func (t *ControllableMockTransport) Read(b []byte) (int, error) {
	t.mu.Lock()
	if t.unplugged {
		t.mu.Unlock()
		return 0, io.EOF
	}
	if t.ioErr != nil {
		err := t.ioErr
		t.mu.Unlock()
		return 0, err
	}
	pipe := t.pipe
	t.mu.Unlock()

	return pipe.Read(b)
}

func (t *ControllableMockTransport) Write(b []byte) (int, error) {
	t.mu.Lock()
	if t.unplugged {
		t.mu.Unlock()
		return 0, fmt.Errorf("input/output error")
	}
	if t.ioErr != nil {
		err := t.ioErr
		t.mu.Unlock()
		return 0, err
	}
	pipe := t.pipe
	t.mu.Unlock()

	return pipe.Write(b)
}

func (t *ControllableMockTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closeCount++
	if t.pipe != nil {
		_ = t.pipe.Close()
	}
	return nil
}

func (t *ControllableMockTransport) SetDeadline(tm time.Time) error {
	return nil
}

func (t *ControllableMockTransport) SetUnplugged(unplugged bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.unplugged = unplugged
	if unplugged && t.pipe != nil {
		_ = t.pipe.Close()
	}
}

func (t *ControllableMockTransport) SetIOError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ioErr = err
}

// MockPortRegistry manages virtual serial port lifecycle across simulated physical ports
type MockPortRegistry struct {
	mu        sync.Mutex
	ports     map[string]*ControllableMockTransport
	openCalls map[string]int
}

func newMockPortRegistry() *MockPortRegistry {
	return &MockPortRegistry{
		ports:     make(map[string]*ControllableMockTransport),
		openCalls: make(map[string]int),
	}
}

func (r *MockPortRegistry) RegisterPort(port string, server *modbus.MockModbusServer) *ControllableMockTransport {
	r.mu.Lock()
	defer r.mu.Unlock()
	tr := newControllableTransport(port, server)
	r.ports[port] = tr
	return tr
}

func (r *MockPortRegistry) RemovePort(port string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tr, exists := r.ports[port]; exists {
		tr.SetUnplugged(true)
		delete(r.ports, port)
	}
}

func (r *MockPortRegistry) Opener(cfg *model.DeviceConnection) (modbus.SerialTransport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	port := modbus.ResolvePortAddress(cfg.SerialPort)
	if port == "" {
		port = cfg.SerialPort
	}
	r.openCalls[port]++

	tr, exists := r.ports[port]
	if !exists || tr.unplugged {
		return nil, fmt.Errorf("open %s: no such file or directory", port)
	}
	// Return a fresh pipe for new open
	tr.pipe = modbus.NewMockSerialPipe(tr.server)
	return tr, nil
}

// 1. USB serial disconnect while polling & 2. Serial I/O failure followed by port closure
func TestHotplug_SerialDisconnectWhilePolling(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 42)

	portName := "/dev/ttyUSB0"
	tr := reg.RegisterPort(portName, server)

	adapter := modbus.NewModbusRTUAdapterWithOpener(&model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: portName,
		BaudRate:   9600,
		Timeout:    1000,
		SlaveID:    1,
	}, reg.Opener)

	ctx := context.Background()
	if err := adapter.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// 1. Verify successful communication
	resp, err := adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err != nil || len(resp.Data) < 2 {
		t.Fatalf("Initial ReadRegisters failed: %v", err)
	}

	// 2. Physical unplug happens!
	tr.SetUnplugged(true)
	reg.RemovePort(portName)

	// 3. Next read detects I/O failure and closes invalid port handle
	_, err = adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err == nil {
		t.Fatalf("Expected ReadRegisters to fail after USB disconnect, got nil")
	}

	// Verify port handle was closed and adapter marked with error
	status := adapter.GetStatus()
	if status.State != modbus.StateError {
		t.Errorf("Expected adapter state ERROR, got %v", status.State)
	}
	if tr.closeCount == 0 {
		t.Errorf("Expected invalid serial transport handle to be closed, closeCount is 0")
	}
}

// 3. Port absent during reconnect
func TestHotplug_PortAbsentDuringReconnect(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	portName := "/dev/ttyUSB_MISSING"

	adapter := modbus.NewModbusRTUAdapterWithOpener(&model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: portName,
		BaudRate:   9600,
		Timeout:    500,
		SlaveID:    1,
	}, reg.Opener)

	ctx := context.Background()
	err := adapter.Connect(ctx)
	if err == nil {
		t.Fatalf("Expected Connect to fail when port is absent, got nil")
	}

	if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "failed to open") {
		t.Errorf("Unexpected error message: %v", err)
	}

	status := adapter.GetStatus()
	if status.State != modbus.StateError {
		t.Errorf("Expected adapter state ERROR, got %v", status.State)
	}
}

// 4. Same device path returning after reinsertion & 6. Reopen and successful polling recovery
func TestHotplug_SameDevicePathReinsertionAndRecovery(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 1234)

	portName := "/dev/ttyUSB0"
	tr := reg.RegisterPort(portName, server)

	adapter := modbus.NewModbusRTUAdapterWithOpener(&model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: portName,
		BaudRate:   9600,
		Timeout:    500,
		SlaveID:    1,
	}, reg.Opener)

	ctx := context.Background()
	if err := adapter.Connect(ctx); err != nil {
		t.Fatalf("Initial Connect failed: %v", err)
	}

	// 1. Initial read OK
	resp, err := adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("Initial read failed: %v", err)
	}
	if val := int(resp.Data[0])<<8 | int(resp.Data[1]); val != 1234 {
		t.Fatalf("Expected value 1234, got %d", val)
	}

	// 2. Unplug USB
	tr.SetUnplugged(true)
	reg.RemovePort(portName)

	_, err = adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err == nil {
		t.Fatalf("Expected read to fail while unplugged")
	}

	// 3. User plugs the exact same USB device back in
	server.SetHoldingRegister(0, 5678)
	newTr := reg.RegisterPort(portName, server)

	// 4. Next read must automatically reopen port and recover without restart!
	resp2, err2 := adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err2 != nil {
		t.Fatalf("Read after reinsertion failed: %v", err2)
	}
	if val2 := int(resp2.Data[0])<<8 | int(resp2.Data[1]); val2 != 5678 {
		t.Fatalf("Expected recovered value 5678, got %d", val2)
	}

	status := adapter.GetStatus()
	if status.State != modbus.StateConnected {
		t.Errorf("Expected adapter state CONNECTED after recovery, got %v", status.State)
	}
	_ = newTr
}

// 5. Device path changing after reinsertion (e.g. ttyUSB0 -> ttyUSB1 jump)
func TestHotplug_DevicePathChangingAfterReinsertion(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 9999)

	port1 := "/dev/ttyUSB0"
	port2 := "/dev/ttyUSB1"
	tr1 := reg.RegisterPort(port1, server)

	connCfg := &model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: port1,
		BaudRate:   9600,
		Timeout:    500,
		SlaveID:    1,
	}

	adapter := modbus.NewModbusRTUAdapterWithOpener(connCfg, reg.Opener)
	ctx := context.Background()

	if err := adapter.Connect(ctx); err != nil {
		t.Fatalf("Connect on port1 failed: %v", err)
	}

	// Unplug port1
	tr1.SetUnplugged(true)
	reg.RemovePort(port1)

	_, _ = adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})

	// Reinsertion causes device to appear at port2
	connCfg.SerialPort = port2
	reg.RegisterPort(port2, server)

	// Read should cleanly migrate to port2 without hanging
	resp, err := adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("Read on migrated port2 failed: %v", err)
	}
	if val := int(resp.Data[0])<<8 | int(resp.Data[1]); val != 9999 {
		t.Fatalf("Expected value 9999 from migrated port, got %d", val)
	}
}

// 7. Diagnostic lock released after a failed test
func TestHotplug_DiagnosticLockReleasedAfterFailedTest(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	dev := &model.Device{
		ID:         701,
		DeviceCode: "TEST-DEV-701",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "/dev/ttyUSB_FAIL",
			Timeout:    500,
			SlaveID:    1,
		},
	}

	mgr := communication.NewConnectionManager(nil, func(cfg *model.DeviceConnection) (communication.ProtocolAdapter, error) {
		return modbus.NewModbusRTUAdapterWithOpener(cfg, reg.Opener), nil
	})

	ctx := context.Background()
	// Test on nonexistent port -> must fail with SERIAL_OPEN_ERROR
	diag1 := mgr.DiagnosticTest(ctx, dev)
	if diag1.Success || diag1.Connected {
		t.Fatalf("Expected diagnostic to fail on offline port, got success")
	}
	if diag1.Code != modbus.DiagSerialOpenError {
		t.Errorf("Expected DiagSerialOpenError, got %v", diag1.Code)
	}

	// Verify diagnostic lock is NOT stuck
	if mgr.IsDiagnosticActive(dev.ID) {
		t.Fatalf("Diagnostic lock remained stuck after failed test!")
	}

	// Immediate second diagnostic test must execute and not return BUSY
	diag2 := mgr.DiagnosticTest(ctx, dev)
	if diag2.Code == modbus.DiagBusy {
		t.Fatalf("Second diagnostic returned BUSY: lock was not released!")
	}
}

// 8. Diagnostic timeout followed by a successful new test
func TestHotplug_DiagnosticTimeoutFollowedBySuccessfulTest(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	server.SetMode(modbus.SimModeTimeout, 0, 3*time.Second) // Delay longer than test timeout

	portName := "/dev/ttyUSB_SLOW"
	reg.RegisterPort(portName, server)

	dev := &model.Device{
		ID:         801,
		DeviceCode: "TEST-DEV-801",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: portName,
			Timeout:    300, // 300ms timeout
			SlaveID:    1,
		},
	}

	mgr := communication.NewConnectionManager(nil, func(cfg *model.DeviceConnection) (communication.ProtocolAdapter, error) {
		return modbus.NewModbusRTUAdapterWithOpener(cfg, reg.Opener), nil
	})

	ctx := context.Background()
	diag1 := mgr.DiagnosticTest(ctx, dev)
	if diag1.Success {
		t.Fatalf("Expected timeout failure, got success")
	}

	// Normal server mode restored
	server.SetMode(modbus.SimModeNormal, 0, 0)
	server.SetHoldingRegister(0, 777)

	diag2 := mgr.DiagnosticTest(ctx, dev)
	if !diag2.Success {
		t.Fatalf("Expected second test to succeed after timeout recovery, got: %s", diag2.Message)
	}
	if mgr.IsDiagnosticActive(dev.ID) {
		t.Errorf("Diagnostic lock should be free after test completion")
	}
}

// 9. Concurrent Test Link requests serialized or rejected predictably
func TestHotplug_ConcurrentTestLinkRequests(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	server.SetMode(modbus.SimModeNormal, 0, 50*time.Millisecond)

	portName := "/dev/ttyUSB_CONCUR"
	reg.RegisterPort(portName, server)

	dev := &model.Device{
		ID:         901,
		DeviceCode: "TEST-DEV-901",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: portName,
			Timeout:    1000,
			SlaveID:    1,
		},
	}

	mgr := communication.NewConnectionManager(nil, func(cfg *model.DeviceConnection) (communication.ProtocolAdapter, error) {
		return modbus.NewModbusRTUAdapterWithOpener(cfg, reg.Opener), nil
	})

	var wg sync.WaitGroup
	results := make([]*communication.DiagnosticResult, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = mgr.DiagnosticTest(context.Background(), dev)
		}(i)
	}
	wg.Wait()

	busyCount := 0
	successCount := 0
	for _, res := range results {
		if res.Code == modbus.DiagBusy {
			busyCount++
		} else if res.Success {
			successCount++
		}
	}

	if successCount == 0 {
		t.Errorf("Expected at least 1 diagnostic test to succeed, got 0")
	}
	if busyCount == 0 && successCount < 2 {
		t.Errorf("Expected concurrent attempts to be either serialized or return DiagBusy")
	}
	if mgr.IsDiagnosticActive(dev.ID) {
		t.Errorf("Diagnostic lock must be free after all goroutines finish")
	}
}

// 10. Reconnect racing with diagnostic execution
func TestHotplug_ReconnectRacingWithDiagnostic(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	portName := "/dev/ttyUSB_RACE"
	reg.RegisterPort(portName, server)

	dev := &model.Device{
		ID:         1001,
		DeviceCode: "TEST-DEV-1001",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: portName,
			Timeout:    500,
			SlaveID:    1,
		},
	}

	mgr := communication.NewConnectionManager(nil, func(cfg *model.DeviceConnection) (communication.ProtocolAdapter, error) {
		return modbus.NewModbusRTUAdapterWithOpener(cfg, reg.Opener), nil
	})

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		reconnCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = mgr.ReconnectDevice(reconnCtx, dev)
	}()

	var diagRes *communication.DiagnosticResult
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // Ensure reconnect started
		diagRes = mgr.DiagnosticTest(context.Background(), dev)
	}()

	wg.Wait()

	if diagRes == nil {
		t.Fatalf("Diagnostic result was nil")
	}
	// Diagnostic either detected reconnection and reported busy gracefully or succeeded
	if diagRes.Code != modbus.DiagBusy && !diagRes.Success {
		t.Errorf("Unexpected diagnostic result during reconnect race: Code=%v, Msg=%s", diagRes.Code, diagRes.Message)
	}
}

// 11. Repeated unplug/replug cycles without goroutine or handle leaks
func TestHotplug_RepeatedUnplugReplugCyclesWithoutLeaks(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	portName := "/dev/ttyUSB_CYCLE"
	tr := reg.RegisterPort(portName, server)

	adapter := modbus.NewModbusRTUAdapterWithOpener(&model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: portName,
		BaudRate:   9600,
		Timeout:    200,
		SlaveID:    1,
	}, reg.Opener)

	ctx := context.Background()
	_ = adapter.Connect(ctx)

	baseGoroutines := runtime.NumGoroutine()

	// Execute 10 repeated disconnect and reconnect cycles
	for cycle := 1; cycle <= 10; cycle++ {
		// 1. Unplug
		tr.SetUnplugged(true)
		_, _ = adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
			SlaveID:         1,
			FunctionCode:    modbus.FunctionReadHoldingRegisters,
			StartingAddress: 0,
			Quantity:        1,
		})

		// 2. Replug
		tr.SetUnplugged(false)
		_, err := adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
			SlaveID:         1,
			FunctionCode:    modbus.FunctionReadHoldingRegisters,
			StartingAddress: 0,
			Quantity:        1,
		})
		if err != nil {
			t.Fatalf("Cycle %d: read failed after reinsertion: %v", cycle, err)
		}
	}

	endGoroutines := runtime.NumGoroutine()
	// Allow small delta for GC / runtime workers
	if endGoroutines-baseGoroutines > 5 {
		t.Errorf("Potential goroutine leak detected across 10 cycles: base=%d, end=%d",
			baseGoroutines, endGoroutines)
	}
}

// 12. Two devices operating independently while one disconnects
func TestHotplug_TwoDevicesIndependentIsolation(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	serverA := modbus.NewMockModbusServer()
	serverA.SetHoldingRegister(0, 111)
	serverB := modbus.NewMockModbusServer()
	serverB.SetHoldingRegister(0, 222)

	portA := "/dev/ttyUSB_DEV_A"
	portB := "/dev/ttyUSB_DEV_B"
	trA := reg.RegisterPort(portA, serverA)
	_ = reg.RegisterPort(portB, serverB)

	devA := &model.Device{
		ID:         1201,
		DeviceCode: "DEV-A",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: portA,
			Timeout:    500,
			SlaveID:    1,
		},
	}
	devB := &model.Device{
		ID:         1202,
		DeviceCode: "DEV-B",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: portB,
			Timeout:    500,
			SlaveID:    1,
		},
	}

	mgr := communication.NewConnectionManager(nil, func(cfg *model.DeviceConnection) (communication.ProtocolAdapter, error) {
		return modbus.NewModbusRTUAdapterWithOpener(cfg, reg.Opener), nil
	})

	ctx := context.Background()
	_ = mgr.ConnectDevice(ctx, devA)
	_ = mgr.ConnectDevice(ctx, devB)

	// Unplug Device A's cable
	trA.SetUnplugged(true)

	// Read Device A -> fails
	_, errA := mgr.ExecuteWithRetry(ctx, devA, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if errA == nil {
		t.Fatalf("Expected Device A read to fail while unplugged")
	}

	// Read Device B -> must succeed completely uninterrupted!
	respB, errB := mgr.ExecuteWithRetry(ctx, devB, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if errB != nil {
		t.Fatalf("Device B read failed when Device A was unplugged: %v", errB)
	}
	if valB := int(respB.Data[0])<<8 | int(respB.Data[1]); valB != 222 {
		t.Fatalf("Expected value 222 from Device B, got %d", valB)
	}

	// Verify Device B health is connected and Device A is degraded/error
	healthB := mgr.GetDeviceHealth(devB.ID)
	if healthB.State != communication.StateConnected {
		t.Errorf("Expected Device B to remain CONNECTED, got %v", healthB.State)
	}
}

// 13. Correct communication status and retry counters
func TestHotplug_StatusAndRetryCounters(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	portName := "/dev/ttyUSB_STATUS"
	tr := reg.RegisterPort(portName, server)

	dev := &model.Device{
		ID:         1301,
		DeviceCode: "DEV-STATUS",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: portName,
			Timeout:    200,
			RetryCount: 3,
			SlaveID:    1,
		},
	}

	mgr := communication.NewConnectionManager(nil, func(cfg *model.DeviceConnection) (communication.ProtocolAdapter, error) {
		return modbus.NewModbusRTUAdapterWithOpener(cfg, reg.Opener), nil
	})

	ctx := context.Background()
	_ = mgr.ConnectDevice(ctx, dev)

	// Unplug USB
	tr.SetUnplugged(true)

	// Execute with retry
	_, err := mgr.ExecuteWithRetry(ctx, dev, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err == nil {
		t.Fatalf("Expected failure on unplugged port")
	}

	health := mgr.GetDeviceHealth(dev.ID)
	if health.ConsecutiveFailures == 0 {
		t.Errorf("Expected ConsecutiveFailures > 0, got 0")
	}
	if health.LastError == "" {
		t.Errorf("Expected LastError to be recorded, got empty string")
	}
	if health.LastCommFailure == nil {
		t.Errorf("Expected LastCommFailure timestamp to be recorded")
	}

	// Replug USB
	tr.SetUnplugged(false)
	server.SetHoldingRegister(0, 500)

	_, errSuccess := mgr.ExecuteWithRetry(ctx, dev, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if errSuccess != nil {
		t.Fatalf("Expected recovery read to succeed: %v", errSuccess)
	}

	recoveredHealth := mgr.GetDeviceHealth(dev.ID)
	if recoveredHealth.ConsecutiveFailures != 0 {
		t.Errorf("Expected ConsecutiveFailures to reset to 0 after recovery, got %d", recoveredHealth.ConsecutiveFailures)
	}
	if recoveredHealth.ConsecutiveSuccesses == 0 {
		t.Errorf("Expected ConsecutiveSuccesses > 0 after recovery, got 0")
	}
	if recoveredHealth.State != communication.StateConnected {
		t.Errorf("Expected StateConnected after recovery, got %v", recoveredHealth.State)
	}
}
