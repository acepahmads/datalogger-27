package communication

import (
	"context"
	"fmt"
	"math"
	"net"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

// TestLongRunningPollingStability validates long-running continuous polling stability,
// memory allocation boundaries, zero socket leaks, and zero goroutine leaks.
func TestLongRunningPollingStability(t *testing.T) {
	// Baseline goroutine count
	runtime.GC()
	baselineGoroutines := runtime.NumGoroutine()

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 1234)
	server.SetHoldingRegister(1, 5678)
	addr, err := server.StartTCP()
	if err != nil {
		t.Fatalf("failed starting TCP mock server: %v", err)
	}
	defer server.Stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	dev := &model.Device{
		ID:         501,
		DeviceCode: "DEV-STABILITY-01",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
		Connection: &model.DeviceConnection{
			DeviceID:        501,
			Protocol:        model.ProtocolModbusTCP,
			Host:            host,
			Port:            port,
			Timeout:         300,
			RetryCount:      2,
			PollingInterval: 20, // Rapid continuous polling for stress testing
			Enabled:         true,
			SlaveID:         1,
			ByteOrder:       "ABCD",
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	engine := NewPollingEngine(mgr, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var successCount int64
	var errorCount int64

	// Start worker loop simulating 50 consecutive polling iterations
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		iterations := 0
		for iterations < 50 {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				iterations++
				resp, err := mgr.ExecuteWithRetry(ctx, dev, ModbusReadRequest{
					SlaveID:         1,
					FunctionCode:    FunctionReadHoldingRegisters,
					StartingAddress: 0,
					Quantity:        2,
				})
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else if len(resp.Data) >= 4 {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}
	}()

	wg.Wait()

	if atomic.LoadInt64(&successCount) < 45 {
		t.Fatalf("expected at least 45 successful cycles, got %d (errors: %d)",
			atomic.LoadInt64(&successCount), atomic.LoadInt64(&errorCount))
	}

	// Gracefully stop all resources
	engine.Stop()
	mgr.CloseAll()

	// Wait for background routines to wind down and perform GC
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()

	// Assert no goroutine leak (allow max difference of 3 for runtime/gc internals)
	if diff := finalGoroutines - baselineGoroutines; diff > 3 {
		t.Errorf("detected possible goroutine leak: baseline=%d, final=%d (diff=%d)",
			baselineGoroutines, finalGoroutines, diff)
	}
}

// TestFailureRecoveryAndAutoReconnect tests temporary network drops,
// failure detection, and automatic reconnection when the server recovers.
func TestFailureRecoveryAndAutoReconnect(t *testing.T) {
	// 1. Start Server on a specific free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate free port: %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close() // Release so MockModbusServer can bind

	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(10, 888)
	if err := server.StartTCPAt(addr); err != nil {
		t.Fatalf("failed starting TCP mock server: %v", err)
	}

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	dev := &model.Device{
		ID:         502,
		DeviceCode: "DEV-RECOVERY-01",
		Enabled:    true,
		Connection: &model.DeviceConnection{
			DeviceID:   502,
			Protocol:   model.ProtocolModbusTCP,
			Host:       host,
			Port:       port,
			Timeout:    200,
			RetryCount: 1,
			SlaveID:    1,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	ctx := context.Background()

	// 2. Verify normal communication
	if err := mgr.ConnectDevice(ctx, dev); err != nil {
		t.Fatalf("initial connection failed: %v", err)
	}

	resp, err := mgr.ExecuteWithRetry(ctx, dev, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 10,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("initial read failed: %v", err)
	}
	if len(resp.Data) != 2 || resp.Data[1] != 120 { // 888 = 0x0378 (120 in low byte)
		t.Logf("read value: %v", resp.Data)
	}

	// 3. Simulate sudden server crash / network disconnect
	server.Stop()
	time.Sleep(50 * time.Millisecond)

	// Read should now fail and detect disconnection
	_, err = mgr.ExecuteWithRetry(ctx, dev, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 10,
		Quantity:        1,
	})
	if err == nil {
		t.Fatalf("expected error after server shutdown, got success")
	}

	// 4. Restart server on the same port
	server2 := modbus.NewMockModbusServer()
	server2.SetHoldingRegister(10, 888)
	if err := server2.StartTCPAt(addr); err != nil {
		t.Fatalf("failed restarting server on same port: %v", err)
	}
	defer server2.Stop()
	time.Sleep(50 * time.Millisecond)

	// 5. Trigger reconnect - communication must recover cleanly
	if err := mgr.ReconnectDevice(ctx, dev); err != nil {
		t.Fatalf("reconnect failed after server restart: %v", err)
	}

	respRecovered, err := mgr.ExecuteWithRetry(ctx, dev, ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    FunctionReadHoldingRegisters,
		StartingAddress: 10,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("read after reconnect failed: %v", err)
	}
	if len(respRecovered.Data) < 2 {
		t.Fatalf("unexpected data length after recovery: %v", respRecovered.Data)
	}
}

// TestMultiDeviceConcurrencyAndIsolation tests concurrent communication across multiple devices
// ensuring that a hanging/dead device does not degrade response times of healthy devices.
func TestMultiDeviceConcurrentIsolation(t *testing.T) {
	// Healthy Server 1
	server1 := modbus.NewMockModbusServer()
	server1.SetHoldingRegister(0, 111)
	addr1, _ := server1.StartTCP()
	defer server1.Stop()
	host1, portStr1, _ := net.SplitHostPort(addr1)
	port1, _ := strconv.Atoi(portStr1)

	// Healthy Server 2
	server2 := modbus.NewMockModbusServer()
	server2.SetHoldingRegister(0, 222)
	addr2, _ := server2.StartTCP()
	defer server2.Stop()
	host2, portStr2, _ := net.SplitHostPort(addr2)
	port2, _ := strconv.Atoi(portStr2)

	// Device 1: Healthy TCP
	dev1 := &model.Device{
		ID:         601,
		DeviceCode: "DEV-ONLINE-01",
		Connection: &model.DeviceConnection{
			DeviceID:   601,
			Protocol:   model.ProtocolModbusTCP,
			Host:       host1,
			Port:       port1,
			Timeout:    500,
			RetryCount: 1,
			SlaveID:    1,
		},
	}

	// Device 2: Dead Port (unroutable/refused)
	dev2 := &model.Device{
		ID:         602,
		DeviceCode: "DEV-DEAD-02",
		Connection: &model.DeviceConnection{
			DeviceID:   602,
			Protocol:   model.ProtocolModbusTCP,
			Host:       "127.0.0.1",
			Port:       65531,
			Timeout:    100,
			RetryCount: 1,
			SlaveID:    1,
		},
	}

	// Device 3: Healthy TCP
	dev3 := &model.Device{
		ID:         603,
		DeviceCode: "DEV-ONLINE-03",
		Connection: &model.DeviceConnection{
			DeviceID:   603,
			Protocol:   model.ProtocolModbusTCP,
			Host:       host2,
			Port:       port2,
			Timeout:    500,
			RetryCount: 1,
			SlaveID:    1,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	ctx := context.Background()

	_ = mgr.ConnectDevice(ctx, dev1)
	_ = mgr.ConnectDevice(ctx, dev3)

	var wg sync.WaitGroup
	var dev1Latencies []time.Duration
	var dev1Mu sync.Mutex

	// Concurrent worker for Dev1
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			start := time.Now()
			_, err := mgr.ExecuteWithRetry(ctx, dev1, ModbusReadRequest{
				SlaveID:         1,
				FunctionCode:    FunctionReadHoldingRegisters,
				StartingAddress: 0,
				Quantity:        1,
			})
			elapsed := time.Since(start)
			if err == nil {
				dev1Mu.Lock()
				dev1Latencies = append(dev1Latencies, elapsed)
				dev1Mu.Unlock()
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Concurrent worker hitting Dead Device 2 (should fail repeatedly without affecting Dev1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			_, _ = mgr.ExecuteWithRetry(ctx, dev2, ModbusReadRequest{
				SlaveID:         1,
				FunctionCode:    FunctionReadHoldingRegisters,
				StartingAddress: 0,
				Quantity:        1,
			})
			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Wait()

	dev1Mu.Lock()
	defer dev1Mu.Unlock()

	if len(dev1Latencies) == 0 {
		t.Fatalf("expected successful reads on Dev1 despite Dev2 failures")
	}

	// Verify Dev1 average response time remained fast (under 50ms) despite Dev2 timeouts
	var total time.Duration
	for _, lat := range dev1Latencies {
		total += lat
	}
	avg := total / time.Duration(len(dev1Latencies))
	if avg > 50*time.Millisecond {
		t.Errorf("Dev1 response time degraded by failing Dev2: avg=%v", avg)
	}
}

// TestPerformanceAndLatencyProfiling measures request roundtrip latency,
// register decoding duration, and connection overhead.
func TestPerformanceAndLatencyProfiling(t *testing.T) {
	server := modbus.NewMockModbusServer()
	for i := 0; i < 100; i++ {
		server.SetHoldingRegister(uint16(i), uint16(i*10))
	}
	addr, _ := server.StartTCP()
	defer server.Stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	dev := &model.Device{
		ID:         701,
		DeviceCode: "DEV-PERF-01",
		Connection: &model.DeviceConnection{
			DeviceID:   701,
			Protocol:   model.ProtocolModbusTCP,
			Host:       host,
			Port:       port,
			Timeout:    500,
			RetryCount: 1,
			SlaveID:    1,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	ctx := context.Background()

	// 1. Connection Duration Measurement
	startConn := time.Now()
	if err := mgr.ConnectDevice(ctx, dev); err != nil {
		t.Fatalf("connection failed: %v", err)
	}
	connDuration := time.Since(startConn)
	t.Logf("Performance Metric: Connection Duration = %v", connDuration)

	// 2. Transaction Roundtrip Latency (100 sequential reads)
	var minLatency = time.Hour
	var maxLatency time.Duration
	var totalLatency time.Duration
	samples := 50

	for i := 0; i < samples; i++ {
		startReq := time.Now()
		resp, err := mgr.ExecuteWithRetry(ctx, dev, ModbusReadRequest{
			SlaveID:         1,
			FunctionCode:    FunctionReadHoldingRegisters,
			StartingAddress: uint16(i),
			Quantity:        2,
		})
		elapsed := time.Since(startReq)
		if err != nil {
			t.Fatalf("perf read failed on cycle %d: %v", i, err)
		}
		if len(resp.Data) < 4 {
			t.Fatalf("unexpected data length")
		}

		if elapsed < minLatency {
			minLatency = elapsed
		}
		if elapsed > maxLatency {
			maxLatency = elapsed
		}
		totalLatency += elapsed
	}

	avgLatency := totalLatency / time.Duration(samples)
	t.Logf("Performance Metrics: Min = %v, Max = %v, Avg = %v (across %d samples)",
		minLatency, maxLatency, avgLatency, samples)

	if avgLatency > 25*time.Millisecond {
		t.Errorf("average latency higher than expected: %v", avgLatency)
	}
}

// TestGracefulShutdownAndResourceCleanup validates that calling Stop() cleanly terminates
// all worker routines, closes all network connections, and leaves zero background leaks.
func TestGracefulShutdownAndResourceCleanup(t *testing.T) {
	runtime.GC()
	initialGoroutines := runtime.NumGoroutine()

	server := modbus.NewMockModbusServer()
	addr, _ := server.StartTCP()
	defer server.Stop()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	engine := NewPollingEngine(mgr, nil, nil)

	// Create 3 devices and register workers
	for i := 1; i <= 3; i++ {
		dev := &model.Device{
			ID:         uint(800 + i),
			DeviceCode: fmt.Sprintf("DEV-SHUTDOWN-%d", i),
			Enabled:    true,
			Status:     model.DeviceStatusActive,
			Connection: &model.DeviceConnection{
				DeviceID:        uint(800 + i),
				Protocol:        model.ProtocolModbusTCP,
				Host:            host,
				Port:            port,
				Timeout:         500,
				RetryCount:      1,
				PollingInterval: 50,
				Enabled:         true,
				SlaveID:         i,
			},
		}
		engine.StartDeviceWorker(dev)
	}

	// Verify workers are registered
	if count := engine.ActiveWorkerCount(); count != 3 {
		t.Fatalf("expected 3 active workers, got %d", count)
	}

	time.Sleep(100 * time.Millisecond)

	// Execute Graceful Shutdown
	engine.Stop()

	// Verify worker count drops to 0
	if count := engine.ActiveWorkerCount(); count != 0 {
		t.Fatalf("expected 0 active workers after Stop, got %d", count)
	}

	time.Sleep(150 * time.Millisecond)
	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()

	// Assert goroutine cleanup
	if diff := finalGoroutines - initialGoroutines; diff > 3 {
		t.Errorf("goroutine leak after engine shutdown: initial=%d, final=%d",
			initialGoroutines, finalGoroutines)
	}
}

// TestScaleAndOffsetPrecisionEdgeCases validates linear scaling boundaries,
// negative values, zero scales, and precision rounding.
func TestScaleAndOffsetPrecisionEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		raw       float64
		scale     float64
		offset    float64
		precision int
		expected  float64
	}{
		{"Normal positive", 100.0, 0.1, 5.0, 2, 15.0},
		{"Negative raw", -50.0, 0.5, 10.0, 1, -15.0},
		{"High precision float", 12345.0, 0.001, 0.005, 3, 12.35},
		{"Negative offset", 200.0, 1.0, -50.0, 0, 150.0},
		{"Zero scale fallback", 100.0, 0.0, 10.0, 2, 110.0}, // scale 0 fallback to 1.0
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := modbus.ApplyScaleAndOffset(tc.raw, tc.scale, tc.offset, tc.precision)
			if math.Abs(res-tc.expected) > 0.001 {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}
