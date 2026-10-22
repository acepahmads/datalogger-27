package communication

import (
	"context"
	"math"
	"net"
	"strconv"
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

func TestBatchPolling_ExecutionAndEquivalence(t *testing.T) {
	server := modbus.NewMockModbusServer()

	// Register 0 & 1 (40001 & 40002): Float32 230.5 (0x43668000)
	fVal := float32(230.5)
	bits := math.Float32bits(fVal)
	server.SetHoldingRegister(0, uint16((bits>>16)&0xFFFF))
	server.SetHoldingRegister(1, uint16(bits&0xFFFF))

	// Register 2 (40003): Int16 500 with BA byte order -> 0x01F4 swapped = 0xF401
	server.SetHoldingRegister(2, 0xF401)

	// Register 3 (40004): Uint16 1234
	server.SetHoldingRegister(3, 1234)

	addr, err := server.StartTCP()
	if err != nil {
		t.Fatalf("failed starting mock server: %v", err)
	}
	defer server.Stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	dev := &model.Device{
		ID:         401,
		DeviceCode: "DEV-BATCH-01",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
		Connection: &model.DeviceConnection{
			DeviceID:         401,
			Protocol:         model.ProtocolModbusTCP,
			Host:             host,
			Port:             port,
			Timeout:          500,
			RetryCount:       1,
			PollingInterval:  1000,
			Enabled:          true,
			SlaveID:          1,
			ByteOrder:        "ABCD",
			BatchReadEnabled: true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:   0,
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			DeviceID:        401,
			ParameterCode:   "VOLT",
			ParameterName:   "Voltage",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40001,
			DataType:        model.DataTypeFloat32,
			ByteOrder:       "ABCD",
			Scale:           1.0,
			Precision:       1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			DeviceID:        401,
			ParameterCode:   "FREQ",
			ParameterName:   "Frequency",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40003,
			DataType:        model.DataTypeInt16,
			ByteOrder:       "BA",
			Scale:           0.1, // 500 * 0.1 = 50.0
			Precision:       1,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              3,
			DeviceID:        401,
			ParameterCode:   "COUNT",
			ParameterName:   "Counter",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40004,
			DataType:        model.DataTypeUInt16,
			ByteOrder:       "AB",
			Scale:           1.0,
			Precision:       0,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	engine := NewPollingEngine(mgr, nil, nil)
	ctx := context.Background()

	// 1. Read individually first to get baseline reference results
	singleResults := make(map[uint]*CommunicationResult)
	for i := range params {
		res, rErr := engine.ReadParameter(ctx, dev, &params[i])
		if rErr != nil || !res.Success {
			t.Fatalf("Single read baseline failed for %s: %v", params[i].ParameterCode, rErr)
		}
		singleResults[params[i].ID] = res
	}

	// 2. Execute Batch Read via pollDeviceBatch
	ok := engine.pollDeviceBatch(ctx, dev, params)
	if !ok {
		t.Fatalf("pollDeviceBatch returned false; expected true")
	}

	// 3. Verify Batch Plan
	plan := modbus.PlanDeviceBatches(dev, params)
	if len(plan.Jobs) != 1 {
		t.Fatalf("expected exactly 1 batch job, got %d", len(plan.Jobs))
	}
	if plan.Jobs[0].Quantity != 4 { // 2 + 1 + 1 registers
		t.Fatalf("expected batch quantity 4, got %d", plan.Jobs[0].Quantity)
	}

	// 4. Verify Equivalence: compare sliced batch results with single read
	req := modbus.ModbusReadRequest{
		SlaveID:         plan.Jobs[0].SlaveID,
		FunctionCode:    plan.Jobs[0].FunctionCode,
		StartingAddress: plan.Jobs[0].StartingAddress,
		Quantity:        plan.Jobs[0].Quantity,
	}
	batchResp, err := mgr.ExecuteWithRetry(ctx, dev, req)
	if err != nil {
		t.Fatalf("raw batch execution failed: %v", err)
	}

	for _, sp := range plan.Jobs[0].Parameters {
		slicedBytes := batchResp.Data[sp.ByteOffset : sp.ByteOffset+sp.ByteLength]
		batchRes, pErr := engine.processDecodedParameter(
			ctx, dev, sp.Parameter, slicedBytes, plan.Jobs[0].FunctionCode, sp.PDUAddress, sp.RegisterCount, batchResp.ResponseTime,
		)
		if pErr != nil {
			t.Fatalf("failed processing sliced parameter %s: %v", sp.Parameter.ParameterCode, pErr)
		}

		singleRes := singleResults[sp.Parameter.ID]
		if math.Abs(batchRes.ProcessedValue-singleRes.ProcessedValue) > 1e-4 {
			t.Fatalf("Equivalence failure on %s: Batch=%v vs Single=%v",
				sp.Parameter.ParameterCode, batchRes.ProcessedValue, singleRes.ProcessedValue)
		}
		if batchRes.Quality != singleRes.Quality {
			t.Fatalf("Quality mismatch on %s: Batch=%v vs Single=%v",
				sp.Parameter.ParameterCode, batchRes.Quality, singleRes.Quality)
		}
	}
}

func TestBatchPolling_OptOutSingleParameter(t *testing.T) {
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 100)
	server.SetHoldingRegister(1, 200)
	server.SetHoldingRegister(2, 300)

	addr, err := server.StartTCP()
	if err != nil {
		t.Fatalf("failed starting mock server: %v", err)
	}
	defer server.Stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	dev := &model.Device{
		ID:         402,
		DeviceCode: "DEV-OPTOUT-01",
		Enabled:    true,
		Connection: &model.DeviceConnection{
			DeviceID:         402,
			Protocol:         model.ProtocolModbusTCP,
			Host:             host,
			Port:             port,
			Timeout:          500,
			RetryCount:       1,
			PollingInterval:  1000,
			Enabled:          true,
			SlaveID:          1,
			ByteOrder:        "AB",
			BatchReadEnabled: true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:   0,
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			DeviceID:        402,
			ParameterCode:   "P1_AUTO",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40001, // PDU 0
			DataType:        model.DataTypeUInt16,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			DeviceID:        402,
			ParameterCode:   "P2_SINGLE",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40002, // PDU 1 -> explicitly opted out to SINGLE
			DataType:        model.DataTypeUInt16,
			BatchMode:       "SINGLE",
			Enabled:         true,
		},
		{
			ID:              3,
			DeviceID:        402,
			ParameterCode:   "P3_AUTO",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40003, // PDU 2
			DataType:        model.DataTypeUInt16,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	plan := modbus.PlanDeviceBatches(dev, params)
	// P1 and P3 are separated by P2, and maxGap=0, so each of P1, P2, P3 is a single job (3 jobs total)
	if len(plan.Jobs) != 3 {
		t.Fatalf("expected 3 jobs due to gap/isolation, got %d", len(plan.Jobs))
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	engine := NewPollingEngine(mgr, nil, nil)
	ok := engine.pollDeviceBatch(context.Background(), dev, params)
	if !ok {
		t.Fatalf("pollDeviceBatch failed for opt-out params")
	}
}

func TestBatchPolling_ControlledFallbackOnBatchFailure(t *testing.T) {
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 111)
	server.SetHoldingRegister(1, 222)

	addr, err := server.StartTCP()
	if err != nil {
		t.Fatalf("failed starting mock server: %v", err)
	}
	defer server.Stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	dev := &model.Device{
		ID:         403,
		DeviceCode: "DEV-FALLBACK-01",
		Enabled:    true,
		Connection: &model.DeviceConnection{
			DeviceID:         403,
			Protocol:         model.ProtocolModbusTCP,
			Host:             host,
			Port:             port,
			Timeout:          500,
			RetryCount:       0,
			PollingInterval:  1000,
			Enabled:          true,
			SlaveID:          1,
			ByteOrder:        "AB",
			BatchReadEnabled: true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:   0,
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			DeviceID:        403,
			ParameterCode:   "P1",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40001,
			DataType:        model.DataTypeUInt16,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			DeviceID:        403,
			ParameterCode:   "P2",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40002,
			DataType:        model.DataTypeUInt16,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	engine := NewPollingEngine(mgr, nil, nil)

	// Inject a custom hook or simulate batch failure:
	// Set mock server mode to EXCEPTION code 0x02
	server.SetMode(modbus.SimModeException, 0x02, 0)

	// Since mock server returns exception on any read, fallback will also gracefully handle failure without crash or loop
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ok := engine.pollDeviceBatch(ctx, dev, params)
	if ok {
		t.Fatalf("expected pollDeviceBatch to return false on full exception, got true")
	}

	// Verify failure counter incremented in batchFailures map
	plan := modbus.PlanDeviceBatches(dev, params)
	if len(plan.Jobs) > 0 {
		batchKey := "403_" + plan.Jobs[0].ID
		engine.batchFailuresMu.Lock()
		count := engine.batchFailures[batchKey]
		engine.batchFailuresMu.Unlock()
		if count != 1 {
			t.Fatalf("expected batch failure count 1, got %d", count)
		}
	}

	// Now restore server to normal
	server.SetMode(modbus.SimModeNormal, 0, 0)
	ok = engine.pollDeviceBatch(ctx, dev, params)
	if !ok {
		t.Fatalf("expected pollDeviceBatch to succeed after server recovered, got false")
	}

	// Verify failure counter was reset
	if len(plan.Jobs) > 0 {
		batchKey := "403_" + plan.Jobs[0].ID
		engine.batchFailuresMu.Lock()
		count := engine.batchFailures[batchKey]
		engine.batchFailuresMu.Unlock()
		if count != 0 {
			t.Fatalf("expected batch failure count reset to 0, got %d", count)
		}
	}
}

func TestBatchPolling_CircuitBreakerTripping(t *testing.T) {
	dev := &model.Device{
		ID:         404,
		DeviceCode: "DEV-CB-01",
		Enabled:    true,
		Connection: &model.DeviceConnection{
			DeviceID:         404,
			Protocol:         model.ProtocolModbusTCP,
			Host:             "127.0.0.1",
			Port:             50200,
			Timeout:          50,
			RetryCount:       0,
			PollingInterval:  1000,
			Enabled:          true,
			SlaveID:          1,
			ByteOrder:        "AB",
			BatchReadEnabled: true,
			MaxBatchRegisters: 32,
			MaxRegisterGap:   0,
		},
	}

	params := []model.Parameter{
		{
			ID:              1,
			DeviceID:        404,
			ParameterCode:   "P1",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40001,
			DataType:        model.DataTypeUInt16,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
		{
			ID:              2,
			DeviceID:        404,
			ParameterCode:   "P2",
			RegisterType:    "HOLDING_REGISTER",
			RegisterAddress: 40002,
			DataType:        model.DataTypeUInt16,
			BatchMode:       "AUTO",
			Enabled:         true,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	engine := NewPollingEngine(mgr, nil, nil)

	plan := modbus.PlanDeviceBatches(dev, params)
	batchKey := "404_" + plan.Jobs[0].ID

	// Pre-set circuit breaker to 3 consecutive failures
	engine.batchFailuresMu.Lock()
	engine.batchFailures[batchKey] = 3
	engine.batchFailuresMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// When polled, it should bypass batch read and directly execute single reads (incrementing failure counter for fallback tracking)
	_ = engine.pollDeviceBatch(ctx, dev, params)

	engine.batchFailuresMu.Lock()
	count := engine.batchFailures[batchKey]
	engine.batchFailuresMu.Unlock()

	if count != 4 {
		t.Fatalf("expected circuit-broken counter to increment to 4, got %d", count)
	}
}
