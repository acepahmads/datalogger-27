package communication

import (
	"context"
	"encoding/binary"
	"math"
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

func TestFormulaCleanTrend_NoOutageOrRecoveryMarkersAndNoZeroSpike(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()
	ResetSharedAdapterTracker()
	defer ResetSharedAdapterTracker()

	server := modbus.NewMockModbusServer()
	// Encode float32(100.0) into holding registers 0..1 (big-endian ABCD)
	bits := math.Float32bits(100.0)
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], bits)
	server.SetHoldingRegister(0, binary.BigEndian.Uint16(b[0:2]))
	server.SetHoldingRegister(1, binary.BigEndian.Uint16(b[2:4]))

	tracer := NewCommTracer(nil)
	tracer.Start("engineer", 5*time.Minute, 0)
	mon := NewActivityMonitor(nil, nil, tracer)

	mgr := setupMockRTUManager(server)
	mgr.SetActivityMonitor(mon)
	mgr.SetCommTracer(tracer)
	mgr.SetBackoffPolicy(&BackoffPolicy{
		InitialInterval: time.Millisecond,
		MaxInterval:     5 * time.Millisecond,
		Multiplier:      1.0,
	})

	tracker := GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)
	tracker.SetConnectionManager(mgr)

	dev, param := createTestDeviceAndParam("COM_TEST_CLEAN_FORMULA", 1, "pm25", "HOLDING", 0, "FLOAT32")
	dev.ID = 777
	dev.DeviceCode = "DEV-FORMULA-777"
	dev.Connection.Timeout = 80
	dev.Connection.RetryCount = 1
	param.ID = 7701
	param.HoldLastValueEnabled = false
	param.Formula = "value * 1.1 + 2.0"
	dev.Parameters = []model.Parameter{*param}

	pe := NewPollingEngine(mgr, nil, nil)
	pe.SetActivityMonitor(mon)
	tracker.RegisterDevice(dev.ID, dev.DeviceCode, dev.DeviceName, dev.Connection.SerialPort)
	tracker.SetDeviceFormulaActive(dev.ID, true)

	ctx := context.Background()

	// 1. Initial physical read succeeds (100.0 -> 100*1.1 + 2 = 112.0)
	if ok := pe.pollDeviceSingle(ctx, dev, dev.Parameters); !ok {
		t.Fatalf("expected initial pollDeviceSingle to succeed")
	}
	tracker.OnCommunicationSuccess(dev.ID, 10)

	// 2. Sensor enters prolonged timeout (SimModeDrop) across 5 cycles + reconnect + port missing/recovered + freshness check
	server.SetMode(modbus.SimModeDrop, 0, 0)
	for i := 0; i < 5; i++ {
		res, err := pe.ReadParameter(ctx, dev, &dev.Parameters[0])
		if err != nil {
			t.Fatalf("cycle %d: expected ReadParameter to succeed via formula fallback, got err=%v", i, err)
		}
		if res == nil || res.FormulaValue == nil {
			t.Fatalf("cycle %d: expected non-nil FormulaValue", i)
		}
		// Must hold last physical reading (100.0 * 1.1 + 2.0 = 112.0) instead of spiking down to 2.0!
		if math.Abs(*res.FormulaValue-112.0) > 1e-5 {
			t.Fatalf("cycle %d: expected smooth FormulaValue=112.0 without zero spike, got %f", i, *res.FormulaValue)
		}
		if res.Quality != model.QualityGood {
			t.Fatalf("cycle %d: expected QualityGood, got %s", i, res.Quality)
		}

		_ = mgr.ReconnectDevice(ctx, dev)
		tracker.OnCommunicationSuccess(dev.ID, 25)
	}

	// Simulate OS-level USB port flap & periodic freshness sweep while formula is active
	tracker.OnPortMissing(dev.ID, dev.Connection.SerialPort, "simulated cable unplug")
	tracker.OnPortRecovered(dev.ID, dev.Connection.SerialPort)
	tracker.EvaluateAllDevicesFreshness()

	// 3. Sensor physically comes back online (SimModeNormal) — must STILL not emit COMMUNICATION_RECOVERED!
	server.SetMode(modbus.SimModeNormal, 0, 0)
	if ok := pe.pollDeviceSingle(ctx, dev, dev.Parameters); !ok {
		t.Fatalf("expected recovery pollDeviceSingle to succeed")
	}
	tracker.OnCommunicationSuccess(dev.ID, 8)

	// 4. Verify ZERO Trend Analysis outage or recovery events were recorded for this device
	noisyEvents := map[string]bool{
		"USB_PORT_MISSING":        true,
		"SERIAL_OPEN_FAILED":      true,
		"DEVICE_OFFLINE":          true,
		"COMMUNICATION_LOST":      true,
		"COMMUNICATION_TIMEOUT":   true,
		"COMMUNICATION_RECOVERED": true,
		"SERIAL_PORT_REOPENED":    true,
	}

	events, _ := mon.GetEvents(ActivityFilter{DeviceID: dev.ID, Limit: 200})
	for _, ev := range events {
		if noisyEvents[ev.EventName] {
			t.Fatalf("expected ZERO outage/recovery events for formula-backed device, but found %s: %s", ev.EventName, ev.Description)
		}
	}

	// Snapshot must remain ONLINE and fresh throughout
	snap, exists := mon.GetDevicePollingSnapshot(dev.ID)
	if !exists || snap == nil {
		t.Fatalf("expected polling snapshot for device %d", dev.ID)
	}
	if snap.CommunicationStatus != CommOnline || !snap.IsFresh {
		t.Fatalf("expected snapshot ONLINE & fresh, got status=%s isFresh=%v", snap.CommunicationStatus, snap.IsFresh)
	}

	// 5. Verify CommTracer still recorded physical timeout truth for Engineers
	frames := tracer.GetRecentFrames(50, dev.ID)
	foundTimeoutFrame := false
	for _, f := range frames {
		if f.Status != "SUCCESS" {
			foundTimeoutFrame = true
			break
		}
	}
	if !foundTimeoutFrame {
		t.Fatalf("expected CommTracer to retain raw physical timeout diagnostics for engineers")
	}
}

func TestFormulaCleanTrend_NeverConnectedDoesNotCompound(t *testing.T) {
	tracker := NewAnomalyHoldTracker()
	param := &model.Parameter{
		ID:                   8801,
		ParameterCode:        "O3",
		Enabled:              true,
		HoldLastValueEnabled: false,
		Formula:              "value + random(34.10, 34.20)",
	}

	for cycle := 0; cycle < 10; cycle++ {
		_, fVal, _, quality, err := tracker.ProcessReading(param, 0, 0, context.DeadlineExceeded)
		if err != nil {
			t.Fatalf("cycle %d: unexpected error: %v", cycle, err)
		}
		if quality != model.QualityGood {
			t.Fatalf("cycle %d: expected QualityGood, got %s", cycle, quality)
		}
		if fVal == nil || *fVal < 34.10 || *fVal > 34.20 {
			t.Fatalf("cycle %d: expected non-compounding FormulaValue in [34.10, 34.20], got %v", cycle, fVal)
		}
	}
}
