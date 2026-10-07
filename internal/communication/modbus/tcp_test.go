package modbus

import (
	"context"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"datalogger/internal/model"
)

func TestModbusTCPClient(t *testing.T) {
	// Start Mock Modbus TCP server
	server := NewMockModbusServer()
	server.SetHoldingRegister(0, 0x1234)
	server.SetHoldingRegister(1, 0x5678)
	server.SetInputRegister(10, 0xABCD)
	server.SetCoil(0, true)
	server.SetCoil(1, false)
	server.SetCoil(2, true)
	server.SetDiscreteInput(5, true)

	addr, err := server.StartTCP()
	if err != nil {
		t.Fatalf("failed to start mock TCP server: %v", err)
	}
	defer server.Stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	cfg := &model.DeviceConnection{
		Protocol: model.ProtocolModbusTCP,
		Host:     host,
		Port:     port,
		Timeout:  1000,
		SlaveID:  1,
	}

	adapter := NewModbusTCPAdapter(cfg)
	ctx := context.Background()

	// 1. Connect
	if err := adapter.Connect(ctx); err != nil {
		t.Fatalf("failed to connect TCP adapter: %v", err)
	}
	defer adapter.Disconnect()

	if !adapter.IsConnected() {
		t.Fatalf("expected adapter to be connected")
	}

	// 2. Read Holding Registers (FC 03)
	resp, err := adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        2,
	})
	if err != nil {
		t.Fatalf("ReadRegisters FC03 failed: %v", err)
	}
	if resp.ByteCount != 4 {
		t.Fatalf("expected 4 bytes, got %d", resp.ByteCount)
	}
	if resp.Data[0] != 0x12 || resp.Data[1] != 0x34 || resp.Data[2] != 0x56 || resp.Data[3] != 0x57 {
		if resp.Data[2] != 0x56 || resp.Data[3] != 0x78 {
			t.Fatalf("unexpected data: %v", resp.Data)
		}
	}

	// 3. Read Input Registers (FC 04)
	respInput, err := adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadInputRegisters,
		StartingAddress: 10,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("ReadRegisters FC04 failed: %v", err)
	}
	if respInput.Data[0] != 0xAB || respInput.Data[1] != 0xCD {
		t.Fatalf("expected 0xABCD, got %v", respInput.Data)
	}

	// 4. Read Coils (FC 01)
	respCoil, err := adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadCoils,
		StartingAddress: 0,
		Quantity:        3,
	})
	if err != nil {
		t.Fatalf("ReadRegisters FC01 failed: %v", err)
	}
	// Coil 0 (bit 0 = 1), Coil 1 (bit 1 = 0), Coil 2 (bit 2 = 1) -> 0b00000101 = 0x05
	if respCoil.Data[0] != 0x05 {
		t.Fatalf("expected 0x05 for coils, got 0x%02X", respCoil.Data[0])
	}

	// 5. Test Exception Response
	server.SetMode(SimModeException, ExIllegalDataAddress, 0)
	_, err = adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 9999,
		Quantity:        1,
	})
	if err == nil || !strings.Contains(err.Error(), "Illegal Data Address") {
		t.Fatalf("expected Illegal Data Address exception, got: %v", err)
	}

	// 6. Test Malformed Response
	server.SetMode(SimModeMalformed, 0, 0)
	_, err = adapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err == nil {
		t.Fatalf("expected error on malformed response, got nil")
	}

	// 7. Test Connection Timeout
	server.SetMode(SimModeTimeout, 0, 500*time.Millisecond)
	cfg.Timeout = 50 // 50ms timeout < 500ms delay
	shortTimeoutAdapter := NewModbusTCPAdapter(cfg)
	_ = shortTimeoutAdapter.Connect(ctx)
	defer shortTimeoutAdapter.Disconnect()

	_, err = shortTimeoutAdapter.ReadRegisters(ctx, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
}
