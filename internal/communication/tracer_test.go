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

	frames := tracer.GetRecentFrames(10, 0)
	if len(frames) != 1 {
		t.Fatalf("Expected 1 captured frame, got %d", len(frames))
	}

	f := frames[0]
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
	if len(tracer.GetRecentFrames(10, 0)) != 1 {
		t.Errorf("Expected no new frames recorded after deactivation")
	}
}
