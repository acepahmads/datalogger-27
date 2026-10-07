package communication

import (
	"context"
	"net"
	"strconv"
	"sync"
	"testing"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

func TestConnectionManagerLifecycle(t *testing.T) {
	// Start Mock Modbus TCP server
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 42)
	addr, err := server.StartTCP()
	if err != nil {
		t.Fatalf("failed starting TCP mock server: %v", err)
	}
	defer server.Stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	devA := &model.Device{
		ID:         101,
		DeviceCode: "DEV-METER-01",
		Connection: &model.DeviceConnection{
			DeviceID:   101,
			Protocol:   model.ProtocolModbusTCP,
			Host:       host,
			Port:       port,
			Timeout:    500,
			RetryCount: 2,
			SlaveID:    1,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	ctx := context.Background()

	// 1. Connect
	if err := mgr.ConnectDevice(ctx, devA); err != nil {
		t.Fatalf("failed to connect device A: %v", err)
	}

	status, exists := mgr.GetAdapterStatus(devA.ID)
	if !exists || status.State != StateConnected {
		t.Fatalf("expected state CONNECTED, got %v", status.State)
	}

	// 2. Read with retry
	req := ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	}
	resp, err := mgr.ExecuteWithRetry(ctx, devA, req)
	if err != nil {
		t.Fatalf("ExecuteWithRetry failed: %v", err)
	}
	if len(resp.Data) != 2 || resp.Data[1] != 42 {
		t.Fatalf("unexpected data: %v", resp.Data)
	}

	// 3. Reconnect
	if err := mgr.ReconnectDevice(ctx, devA); err != nil {
		t.Fatalf("ReconnectDevice failed: %v", err)
	}

	// 4. Disconnect
	if err := mgr.DisconnectDevice(devA.ID); err != nil {
		t.Fatalf("DisconnectDevice failed: %v", err)
	}
	statusAfter, _ := mgr.GetAdapterStatus(devA.ID)
	if statusAfter.State != StateDisconnected {
		t.Fatalf("expected state DISCONNECTED, got %v", statusAfter.State)
	}
}

func TestDeviceIsolation(t *testing.T) {
	// Server for Device B (Healthy)
	serverB := modbus.NewMockModbusServer()
	serverB.SetHoldingRegister(0, 100)
	addrB, _ := serverB.StartTCP()
	defer serverB.Stop()

	hostB, portStrB, _ := net.SplitHostPort(addrB)
	portB, _ := strconv.Atoi(portStrB)

	// Device A targets an invalid/unreachable port (Failed / Offline)
	devA := &model.Device{
		ID:         201,
		DeviceCode: "DEV-OFFLINE-01",
		Connection: &model.DeviceConnection{
			DeviceID:   201,
			Protocol:   model.ProtocolModbusTCP,
			Host:       "127.0.0.1",
			Port:       65530, // dead port
			Timeout:    50,
			RetryCount: 1,
			SlaveID:    1,
		},
	}

	// Device B targets healthy mock server
	devB := &model.Device{
		ID:         202,
		DeviceCode: "DEV-HEALTHY-02",
		Connection: &model.DeviceConnection{
			DeviceID:   202,
			Protocol:   model.ProtocolModbusTCP,
			Host:       hostB,
			Port:       portB,
			Timeout:    500,
			RetryCount: 1,
			SlaveID:    1,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	ctx := context.Background()

	// Connect both concurrently
	var wg sync.WaitGroup
	var errA, errB error

	wg.Add(2)
	go func() {
		defer wg.Done()
		errA = mgr.ConnectDevice(ctx, devA)
	}()
	go func() {
		defer wg.Done()
		errB = mgr.ConnectDevice(ctx, devB)
	}()
	wg.Wait()

	// Device A MUST fail
	if errA == nil {
		t.Fatalf("expected device A to fail connection")
	}

	// Device B MUST succeed (Device A failure must not block Device B!)
	if errB != nil {
		t.Fatalf("device B was blocked by device A failure: %v", errB)
	}

	// Read from Device B
	respB, err := mgr.ExecuteWithRetry(ctx, devB, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("Device B read failed: %v", err)
	}
	if len(respB.Data) != 2 || respB.Data[1] != 100 {
		t.Fatalf("unexpected data from device B: %v", respB.Data)
	}
}
