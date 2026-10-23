package communication

import (
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

func TestCommTracer_LifecycleAndAutoTimeout(t *testing.T) {
	tracer := NewCommTracer(nil)

	// 1. Initial state must be inactive
	if tracer.IsActive() {
		t.Errorf("Expected tracer to be inactive initially")
	}

	// 2. Start tracer with short timeout
	status := tracer.Start("admin", 100*time.Millisecond, 0)
	if !status.Active {
		t.Fatalf("Expected tracer to be active after Start")
	}
	if !tracer.IsActive() {
		t.Fatalf("IsActive should be true")
	}
	if status.ActivatedBy != "admin" {
		t.Errorf("Expected activated_by 'admin', got '%s'", status.ActivatedBy)
	}

	// 3. Record frame while active
	dev := &model.Device{
		ID:         5,
		DeviceCode: "AQMS-02",
		DeviceName: "Station Ambient Air",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "/dev/ttyUSB0",
			BaudRate:   9600,
			DataBits:   8,
			Parity:     "N",
			StopBits:   1,
		},
	}

	req := modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    3,
		StartingAddress: 20,
		Quantity:        4,
	}

	resp := &modbus.ModbusReadResponse{
		SlaveID:      1,
		FunctionCode: 3,
		ByteCount:    8,
		Data:         []byte{0x41, 0xC8, 0x00, 0x00, 0x41, 0x90, 0x00, 0x00},
		ResponseTime: 75 * time.Millisecond,
		RequestADU:   []byte{0x01, 0x03, 0x00, 0x14, 0x00, 0x04, 0x04, 0x0B},
		ResponseADU:  []byte{0x01, 0x03, 0x08, 0x41, 0xC8, 0x00, 0x00, 0x41, 0x90, 0x00, 0x00, 0x5E, 0xD2},
		CRCPassed:    true,
	}

	tracer.Record(dev, req, resp, nil, 75)

	req2 := modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    3,
		StartingAddress: 30,
		Quantity:        2,
	}
	resp2 := &modbus.ModbusReadResponse{
		SlaveID:      1,
		FunctionCode: 3,
		ByteCount:    4,
		Data:         []byte{0x00, 0x01, 0x00, 0x02},
		ResponseTime: 40 * time.Millisecond,
		RequestADU:   []byte{0x01, 0x03, 0x00, 0x1E, 0x00, 0x02, 0xA4, 0x0D},
		ResponseADU:  []byte{0x01, 0x03, 0x04, 0x00, 0x01, 0x00, 0x02, 0x2B, 0x85},
		CRCPassed:    true,
	}
	tracer.Record(dev, req2, resp2, nil, 40)

	frames := tracer.GetRecentFrames(10, 0)
	if len(frames) != 2 {
		t.Fatalf("Expected 2 captured frames, got %d", len(frames))
	}

	// Verify frames[0] is newest frame (StartingAddress 30, DisplayAddress 40031)
	if frames[0].DisplayAddress != 40031 {
		t.Errorf("Expected newest frame at index 0 (DisplayAddress 40031), got %d", frames[0].DisplayAddress)
	}

	f := frames[1]
	if f.DeviceCode != "AQMS-02" {
		t.Errorf("Expected device code AQMS-02, got %s", f.DeviceCode)
	}
	if f.TXHex != "01 03 00 14 00 04 04 0B" {
		t.Errorf("Unexpected TXHex: %s", f.TXHex)
	}
	if f.RXHex != "01 03 08 41 C8 00 00 41 90 00 00 5E D2" {
		t.Errorf("Unexpected RXHex: %s", f.RXHex)
	}
	if f.DisplayAddress != 40021 {
		t.Errorf("Expected display address 40021, got %d", f.DisplayAddress)
	}
	if f.Status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", f.Status)
	}

	// 4. Wait for auto-timeout to expire
	time.Sleep(150 * time.Millisecond)

	if tracer.IsActive() {
		t.Errorf("Expected tracer to be automatically deactivated after timeout")
	}

	// 5. Recording while inactive should be ignored
	tracer.Record(dev, req, resp, nil, 75)
	if len(tracer.GetRecentFrames(10, 0)) != 2 {
		t.Errorf("Expected no new frames recorded after deactivation (count should remain 2)")
	}
}

func TestCommTracer_DecodedAndConvertedValues(t *testing.T) {
	tracer := NewCommTracer(nil)
	tracer.Start("admin", 10*time.Minute, 0)

	dev := &model.Device{
		ID:         10,
		DeviceCode: "AQMS-01",
		DeviceName: "Station Central",
		Connection: &model.DeviceConnection{
			Protocol:  model.ProtocolModbusRTU,
			ByteOrder: "ABCD",
		},
		Parameters: []model.Parameter{
			{
				ID:              101,
				DeviceID:        10,
				ParameterCode:   "pm2.5",
				ParameterName:   "Particulate Matter 2.5",
				RegisterAddress: 40021, // PDU 20
				RegisterType:    "HOLDING_REGISTER",
				DataType:        model.DataTypeFloat32,
				Precision:       2,
				Scale:           1.0,
				Unit:            "µg/m³",
				Enabled:         true,
			},
			{
				ID:              102,
				DeviceID:        10,
				ParameterCode:   "pm10",
				ParameterName:   "Particulate Matter 10",
				RegisterAddress: 40023, // PDU 22
				RegisterType:    "HOLDING_REGISTER",
				DataType:        model.DataTypeFloat32,
				Precision:       2,
				Scale:           1.0,
				Formula:         "x * 1.5", // Setup custom formula
				Unit:            "µg/m³",
				Enabled:         true,
			},
		},
	}

	// 0x41C80000 = 25.0f, 0x41A00000 = 20.0f
	req := modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    3,
		StartingAddress: 20,
		Quantity:        4,
	}
	resp := &modbus.ModbusReadResponse{
		SlaveID:      1,
		FunctionCode: 3,
		ByteCount:    8,
		Data:         []byte{0x41, 0xC8, 0x00, 0x00, 0x41, 0xA0, 0x00, 0x00},
		ResponseTime: 50 * time.Millisecond,
		RequestADU:   []byte{0x01, 0x03, 0x00, 0x14, 0x00, 0x04, 0x04, 0x0B},
		ResponseADU:  []byte{0x01, 0x03, 0x08, 0x41, 0xC8, 0x00, 0x00, 0x41, 0xA0, 0x00, 0x00, 0x00, 0x00},
	}

	tracer.Record(dev, req, resp, nil, 50)
	frames := tracer.GetRecentFrames(1, 0)
	if len(frames) != 1 {
		t.Fatalf("Expected 1 frame, got %d", len(frames))
	}

	f := frames[0]
	if len(f.DecodedValues) != 2 {
		t.Fatalf("Expected 2 decoded values, got %d: %+v", len(f.DecodedValues), f.DecodedValues)
	}

	// First parameter: PM2.5 = 25.00
	if f.DecodedValues[0].ParameterCode != "pm2.5" || f.DecodedValues[0].Formatted != "25.00" {
		t.Errorf("Unexpected DecodedValues[0]: %+v", f.DecodedValues[0])
	}
	if f.DecodedValues[0].Unit != "µg/m³" {
		t.Errorf("Expected unit µg/m³, got '%s'", f.DecodedValues[0].Unit)
	}

	// Second parameter: PM10 raw 20.0, formula x * 1.5 = 30.00
	if f.DecodedValues[1].ParameterCode != "pm10" || f.DecodedValues[1].Formatted != "30.00" {
		t.Errorf("Unexpected DecodedValues[1]: %+v", f.DecodedValues[1])
	}
	if f.DecodedValues[1].Value != 30.0 {
		t.Errorf("Expected formula calculated value 30.0, got %f", f.DecodedValues[1].Value)
	}

	// ConvertedValues string summary
	expectedStr := "pm2.5: 25.00 µg/m³, pm10: 30.00 µg/m³"
	if f.ConvertedValues != expectedStr {
		t.Errorf("Expected ConvertedValues '%s', got '%s'", expectedStr, f.ConvertedValues)
	}
}
