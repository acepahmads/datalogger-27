package modbus

import (
	"context"
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
