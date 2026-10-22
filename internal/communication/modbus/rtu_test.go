package modbus

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"datalogger/internal/model"
)

func TestModbusRTUMaster(t *testing.T) {
	// Initialize mock server
	server := NewMockModbusServer()
	server.SetHoldingRegister(100, 0x1122)
	server.SetInputRegister(200, 0x3344)
	server.SetCoil(10, true)

	pipe := NewMockSerialPipe(server)

	cfg := &model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: "COM3",
		BaudRate:   9600,
		Timeout:    1000,
		SlaveID:    1,
	}

	// Custom opener using MockSerialPipe
	opener := func(c *model.DeviceConnection) (SerialTransport, error) {
		return pipe, nil
	}

	adapter := NewModbusRTUAdapterWithOpener(cfg, opener)
	ctx := context.Background()

	// 1. Connect
	if err := adapter.Connect(ctx); err != nil {
		t.Fatalf("failed to connect RTU adapter: %v", err)
	}
	defer adapter.Disconnect()

	// 2. Read Holding Registers (FC 03)
	resp, err := adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 100,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("ReadRegisters FC03 failed: %v", err)
	}
	if resp.Data[0] != 0x11 || resp.Data[1] != 0x22 {
		t.Fatalf("expected [0x11, 0x22], got %v", resp.Data)
	}

	// 2b. Read Input Registers (FC 04)
	respFC04, err := adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadInputRegisters,
		StartingAddress: 200,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("ReadRegisters FC04 failed: %v", err)
	}
	if respFC04.Data[0] != 0x33 || respFC04.Data[1] != 0x44 {
		t.Fatalf("expected [0x33, 0x44], got %v", respFC04.Data)
	}

	// 2c. Read Coils (FC 01)
	respFC01, err := adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadCoils,
		StartingAddress: 10,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("ReadRegisters FC01 failed: %v", err)
	}
	if len(respFC01.Data) == 0 || (respFC01.Data[0]&0x01) == 0 {
		t.Fatalf("expected coil bit set, got %v", respFC01.Data)
	}

	// 2d. Read Discrete Inputs (FC 02)
	server.SetDiscreteInput(5, true)
	respFC02, err := adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadDiscreteInputs,
		StartingAddress: 5,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("ReadRegisters FC02 failed: %v", err)
	}
	if len(respFC02.Data) == 0 || (respFC02.Data[0]&0x01) == 0 {
		t.Fatalf("expected discrete input bit set, got %v", respFC02.Data)
	}

	// 3. Test CRC Error Detection
	server.SetMode(SimModeCRCError, 0, 0)
	_, err = adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 100,
		Quantity:        1,
	})
	if err == nil || !strings.Contains(err.Error(), "CRC16") {
		t.Fatalf("expected CRC16 error, got %v", err)
	}

	// 4. Test Exception Response
	server.SetMode(SimModeException, ExIllegalDataAddress, 0)
	_, err = adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 9999,
		Quantity:        1,
	})
	if err == nil || !strings.Contains(err.Error(), "Illegal Data Address") {
		t.Fatalf("expected Illegal Data Address exception, got %v", err)
	}

	// 5. Test Malformed Response
	server.SetMode(SimModeMalformed, 0, 0)
	_, err = adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 100,
		Quantity:        1,
	})
	if err == nil {
		t.Fatalf("expected error on malformed RTU response, got nil")
	}

	// 6. Test Timeout
	server.SetMode(SimModeTimeout, 0, 200*time.Millisecond)
	cfg.Timeout = 20
	shortAdapter := NewModbusRTUAdapterWithOpener(cfg, opener)
	_ = shortAdapter.Connect(ctx)
	defer shortAdapter.Disconnect()

	// In memory pipe simulates timeout by sleeping longer than deadline
	_, err = shortAdapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 100,
		Quantity:        1,
	})
	// Should fail or timeout
	if err == nil {
		t.Fatalf("expected timeout or read error, got nil")
	}
}

// TestSharedSerialBusMultiDeviceRecovery verifies that when Dev 1 and Dev 2 share a port,
// Dev 2 (even in StateError) can read registers by attaching to the already open bus without
// attempting to re-open the OS serial handle (which would fail with busy/access denied).
func TestSharedSerialBusMultiDeviceRecovery(t *testing.T) {
	ResetSharedBuses()
	defer ResetSharedBuses()

	server := NewMockModbusServer()
	server.SetHoldingRegister(10, 100)
	server.SetHoldingRegister(20, 200)

	pipe := NewMockSerialPipe(server)
	openerCalls := 0
	opener := func(c *model.DeviceConnection) (SerialTransport, error) {
		openerCalls++
		if openerCalls > 1 {
			// Simulate OS error when port is already opened by another adapter
			return nil, fmt.Errorf("open COM3: Access is denied")
		}
		return pipe, nil
	}

	cfg1 := &model.DeviceConnection{DeviceID: 1, Protocol: model.ProtocolModbusRTU, SerialPort: "COM3", SlaveID: 1, Timeout: 1000}
	cfg2 := &model.DeviceConnection{DeviceID: 2, Protocol: model.ProtocolModbusRTU, SerialPort: "COM3", SlaveID: 1, Timeout: 1000}

	adapter1 := NewModbusRTUAdapterWithOpener(cfg1, opener)
	adapter2 := NewModbusRTUAdapterWithOpener(cfg2, opener)

	ctx := context.Background()

	// 1. Dev 1 connects and opens the shared bus
	if err := adapter1.Connect(ctx); err != nil {
		t.Fatalf("adapter1 connect failed: %v", err)
	}
	defer adapter1.Disconnect()

	// 2. Dev 2 is simulated in StateError / uninitialized (did not call Connect yet)
	adapter2.status.State = StateError

	// 3. Dev 2 reads registers directly - it MUST reuse adapter1's open bus without calling opener again!
	resp2, err2 := adapter2.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 20,
		Quantity:        1,
	})
	if err2 != nil {
		t.Fatalf("adapter2 ReadRegisters failed to reuse shared bus: %v", err2)
	}
	if len(resp2.Data) < 2 || resp2.Data[1] != 200 {
		t.Fatalf("adapter2 unexpected data: %v", resp2.Data)
	}

	if openerCalls != 1 {
		t.Fatalf("expected exactly 1 opener call across both devices, got %d", openerCalls)
	}
	if adapter2.GetStatus().State != StateConnected {
		t.Fatalf("expected adapter2 state to be CONNECTED, got %v", adapter2.GetStatus().State)
	}
}

