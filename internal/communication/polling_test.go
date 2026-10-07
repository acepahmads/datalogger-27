package communication

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"strconv"
	"testing"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

func TestPollingReadParameter(t *testing.T) {
	server := modbus.NewMockModbusServer()
	// Set float 230.5 at holding registers 40001 & 40002 (offset 0 and 1)
	fVal := float32(230.5)
	bits := math.Float32bits(fVal)
	reg0 := uint16((bits >> 16) & 0xFFFF)
	reg1 := uint16(bits & 0xFFFF)
	server.SetHoldingRegister(0, reg0)
	server.SetHoldingRegister(1, reg1)

	// Set frequency 50.0 at input register 30005 (offset 4) as INT16 raw 500 with scale 0.1
	server.SetInputRegister(4, 500)

	addr, _ := server.StartTCP()
	defer server.Stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	dev := &model.Device{
		ID:         301,
		DeviceCode: "DEV-POWER-01",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
		Connection: &model.DeviceConnection{
			DeviceID:        301,
			Protocol:        model.ProtocolModbusTCP,
			Host:            host,
			Port:            port,
			Timeout:         500,
			RetryCount:      1,
			PollingInterval: 1000,
			Enabled:         true,
			SlaveID:         1,
			ByteOrder:       "ABCD",
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	engine := NewPollingEngine(mgr, nil, nil)
	ctx := context.Background()

	// 1. Read FLOAT32 Holding Register Parameter (40001)
	paramVolt := &model.Parameter{
		ID:              1,
		DeviceID:        301,
		ParameterCode:   "VOLT_L1",
		ParameterName:   "Voltage L1",
		RegisterType:    "HOLDING_REGISTER",
		RegisterAddress: 40001,
		DataType:        model.DataTypeFloat32,
		Scale:           1.0,
		Offset:          0.0,
		Precision:       1,
		ByteOrder:       "ABCD",
		Enabled:         true,
	}

	resVolt, err := engine.ReadParameter(ctx, dev, paramVolt)
	if err != nil {
		t.Fatalf("ReadParameter failed: %v", err)
	}
	if !resVolt.Success {
		t.Fatalf("expected successful result, got error: %s", resVolt.ErrorMessage)
	}
	if math.Abs(resVolt.DecodedValue-230.5) > 0.1 {
		t.Fatalf("expected decoded value 230.5, got %v", resVolt.DecodedValue)
	}
	if resVolt.RawHex == "" {
		t.Fatalf("expected formatted RawHex, got empty string")
	}

	// 2. Read Scaled INT16 Input Register Parameter (30005)
	paramFreq := &model.Parameter{
		ID:              2,
		DeviceID:        301,
		ParameterCode:   "FREQ",
		ParameterName:   "Grid Frequency",
		RegisterType:    "INPUT_REGISTER",
		RegisterAddress: 30005,
		DataType:        model.DataTypeInt16,
		Scale:           0.1, // Raw 500 * 0.1 = 50.0 Hz
		Offset:          0.0,
		Precision:       1,
		Enabled:         true,
	}

	resFreq, err := engine.ReadParameter(ctx, dev, paramFreq)
	if err != nil {
		t.Fatalf("ReadParameter for frequency failed: %v", err)
	}
	if resFreq.RawValue != 500 {
		t.Fatalf("expected raw value 500, got %v", resFreq.RawValue)
	}
	if resFreq.DecodedValue != 50.0 {
		t.Fatalf("expected decoded value 50.0, got %v", resFreq.DecodedValue)
	}
}

func init() {
	// Silence unused binary import if needed
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], 0)
}
