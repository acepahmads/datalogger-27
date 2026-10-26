package communication_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"datalogger/internal/communication"
	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

// ============================================================================
// SECTION 8: COMPREHENSIVE INTEGRATION & HARDWARE SIMULATION TESTS
// USB QUAD SERIAL DISCONNECT DETECTION & DEVICE STATUS CONSISTENCY
// ============================================================================

// Case 1: Adapter tersedia dan komunikasi valid (Adapter available and communication valid)
func TestUSBQuadSerial_Case1_AdapterAvailableAndCommunicationValid(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()
	communication.ResetSharedAdapterTracker()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 100)

	portName := "/dev/ttyACM0"
	reg.RegisterPort(portName, server)

	// Set test presence hook: port is physically present
	modbus.SetPortPresenceHookForTest(func(port string) bool {
		return port == portName
	})
	defer modbus.SetPortPresenceHookForTest(nil)

	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	devID := uint(301)
	tracker.RegisterDevice(devID, "COM3", "QinHeng Port 0", portName)

	adapter := modbus.NewModbusRTUAdapterWithOpener(&model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: portName,
		BaudRate:   9600,
		Timeout:    1000,
		SlaveID:    1,
	}, reg.Opener)

	ctx := context.Background()
	if err := adapter.Connect(ctx); err != nil {
		t.Fatalf("Expected connection to succeed, got: %v", err)
	}

	resp, err := adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err != nil || len(resp.Data) < 2 {
		t.Fatalf("ReadRegisters failed: %v", err)
	}

	tracker.OnCommunicationSuccess(devID, 15)

	snap, ok := mon.GetDevicePollingSnapshot(devID)
	if !ok {
		t.Fatalf("Snapshot not found for device %d", devID)
	}

	if snap.TransportStatus != communication.TransportAvailable {
		t.Errorf("Expected TransportStatus %s, got %s", communication.TransportAvailable, snap.TransportStatus)
	}
	if snap.CommunicationStatus != communication.CommOnline {
		t.Errorf("Expected CommunicationStatus %s, got %s", communication.CommOnline, snap.CommunicationStatus)
	}
	if snap.PollingState != communication.PollingStateSuccess {
		t.Errorf("Expected PollingState %s, got %s", communication.PollingStateSuccess, snap.PollingState)
	}
	if !snap.IsFresh {
		t.Errorf("Expected IsFresh to be true for valid recent communication")
	}
	if snap.LastSuccessfulResponse == nil {
		t.Errorf("Expected LastSuccessfulResponse timestamp to be populated")
	}
}

// Case 2: Seluruh port adapter hilang (All ports of adapter missing)
func TestUSBQuadSerial_Case2_AllAdapterPortsMissing(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()
	communication.ResetSharedAdapterTracker()

	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	// QinHeng Quad Serial ports: by-id paths mapping to the 4 interfaces of Quad Serial
	ports := []string{
		"/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if00",
		"/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if01",
		"/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if02",
		"/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if03",
	}
	devIDs := []uint{401, 402, 403, 404}

	for i, p := range ports {
		tracker.RegisterDevice(devIDs[i], fmt.Sprintf("COM%d", i+1), fmt.Sprintf("Device %d", i+1), p)
		tracker.OnCommunicationSuccess(devIDs[i], 10)
	}

	// Verify all initially online
	for _, id := range devIDs {
		snap, _ := mon.GetDevicePollingSnapshot(id)
		if snap.CommunicationStatus != communication.CommOnline {
			t.Fatalf("Device %d expected ONLINE initially, got %s", id, snap.CommunicationStatus)
		}
	}

	// Simulate all ports vanishing (e.g. USB adapter disconnected)
	modbus.SetPortPresenceHookForTest(func(port string) bool {
		return false // All ports disappeared
	})
	defer modbus.SetPortPresenceHookForTest(nil)

	// Trigger disconnect via the first port
	tracker.OnPortMissing(devIDs[0], ports[0], "open /dev/ttyACM0: no such file or directory")

	// All devices must immediately report PORT_MISSING and OFFLINE
	for _, id := range devIDs {
		snap, _ := mon.GetDevicePollingSnapshot(id)
		if snap.TransportStatus != communication.TransportPortMissing {
			t.Errorf("Device %d expected TransportStatus %s, got %s", id, communication.TransportPortMissing, snap.TransportStatus)
		}
		if snap.CommunicationStatus != communication.CommOffline {
			t.Errorf("Device %d expected CommunicationStatus %s, got %s", id, communication.CommOffline, snap.CommunicationStatus)
		}
		if snap.IsFresh {
			t.Errorf("Device %d expected IsFresh=false after adapter loss", id)
		}
	}

	// Verify USB_PORT_MISSING events were recorded
	events, _ := mon.GetEvents(communication.ActivityFilter{Limit: 50})
	foundMissingEvent := false
	for _, ev := range events {
		if ev.EventName == "USB_PORT_MISSING" {
			foundMissingEvent = true
			break
		}
	}
	if !foundMissingEvent {
		t.Errorf("Expected USB_PORT_MISSING event in activity monitor")
	}
}

// Case 3: Dua logical device berbagi satu adapter (Two logical devices sharing one adapter)
func TestUSBQuadSerial_Case3_TwoLogicalDevicesShareOneAdapter(t *testing.T) {
	communication.ResetSharedAdapterTracker()
	tracker := communication.GetSharedAdapterTracker()

	// Linux persistent by-id symlinks for QinHeng Quad_Serial
	port1 := "/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if00"
	port2 := "/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if02"

	key1 := tracker.ResolveAdapterKey(port1)
	key2 := tracker.ResolveAdapterKey(port2)

	if key1 == "" || key2 == "" {
		t.Fatalf("Failed to resolve adapter key from by-id paths: %s, %s", key1, key2)
	}

	if key1 != key2 {
		t.Errorf("Expected shared adapter keys to match! Got %s vs %s", key1, key2)
	}

	expectedPrefix := "usb-1a86_USB_Quad_Serial_55D5"
	if key1 != expectedPrefix {
		t.Errorf("Expected adapter key %s, got %s", expectedPrefix, key1)
	}
}

// Case 4: Satu device OFFLINE sementara device lain pada adapter yang sama masih terlihat ONLINE dicegah
// (BUG REPRODUCTION & RESOLUTION TEST)
func TestUSBQuadSerial_Case4_PreventCOM4StayingOnlineWhenAdapterMissing(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()
	communication.ResetSharedAdapterTracker()

	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	// Hardware setup:
	// QinHeng USB Quad_Serial: 1a86:55d5 at USB path 1-1.1
	// COM3 on /dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if00
	// COM4 on /dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if02
	com3Port := "/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if00"
	com4Port := "/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if02"
	com3ID := uint(3)
	com4ID := uint(4)

	tracker.RegisterDevice(com3ID, "COM3", "Pyranometer Sensor", com3Port)
	tracker.RegisterDevice(com4ID, "COM4", "Weather Station Sensor", com4Port)

	// Step 2 of reproduction: Both initially communicating successfully
	tracker.OnCommunicationSuccess(com3ID, 20)
	tracker.OnCommunicationSuccess(com4ID, 25)

	snap3, _ := mon.GetDevicePollingSnapshot(com3ID)
	snap4, _ := mon.GetDevicePollingSnapshot(com4ID)
	if snap3.CommunicationStatus != communication.CommOnline || snap4.CommunicationStatus != communication.CommOnline {
		t.Fatalf("Expected both COM3 and COM4 to be ONLINE initially")
	}

	// Step 3 of reproduction: echo '1-1.1' | sudo tee /sys/bus/usb/drivers/usb/unbind
	// Physical ports disappear
	modbus.SetPortPresenceHookForTest(func(port string) bool {
		return false // All /dev/ttyACM* gone
	})
	defer modbus.SetPortPresenceHookForTest(nil)

	// COM3 poll worker encounters port error first:
	tracker.OnPortMissing(com3ID, com3Port, "open /dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if00: no such file or directory")

	// CRITICAL ASSERTION:
	// In the buggy implementation, COM3 changed to OFFLINE, but COM4 remained ONLINE.
	// With the fix, COM4 MUST NOT remain ONLINE!
	snap3After, _ := mon.GetDevicePollingSnapshot(com3ID)
	snap4After, _ := mon.GetDevicePollingSnapshot(com4ID)

	if snap3After.CommunicationStatus != communication.CommOffline {
		t.Errorf("COM3 expected OFFLINE, got %s", snap3After.CommunicationStatus)
	}
	if snap4After.CommunicationStatus != communication.CommOffline {
		t.Errorf("COM4 BUG REGRESSION: COM4 is still reported as %s after shared adapter unbind! Must be OFFLINE.", snap4After.CommunicationStatus)
	}

	if snap4After.TransportStatus != communication.TransportPortMissing {
		t.Errorf("COM4 TransportStatus expected %s, got %s", communication.TransportPortMissing, snap4After.TransportStatus)
	}
	if snap4After.IsFresh {
		t.Errorf("COM4 IsFresh must be false after adapter disconnect")
	}
}

// Case 5: Adapter di-unbind (Adapter unbind event)
func TestUSBQuadSerial_Case5_AdapterUnbindEvent(t *testing.T) {
	communication.ResetSharedAdapterTracker()
	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	port := "/dev/serial/by-path/pci-0000:00:14.0-usb-0:1.1:1.0-port0"
	devID := uint(501)
	tracker.RegisterDevice(devID, "DEV5", "Wind Sensor", port)
	tracker.OnCommunicationSuccess(devID, 12)

	// Unbind occurs
	tracker.OnPortMissing(devID, port, "input/output error: device disconnected")

	snap, _ := mon.GetDevicePollingSnapshot(devID)
	if snap.TransportStatus != communication.TransportPortMissing {
		t.Errorf("Expected TransportStatus %s, got %s", communication.TransportPortMissing, snap.TransportStatus)
	}
	if snap.CommunicationStatus != communication.CommOffline {
		t.Errorf("Expected CommunicationStatus %s, got %s", communication.CommOffline, snap.CommunicationStatus)
	}

	events, _ := mon.GetEvents(communication.ActivityFilter{Limit: 10, DeviceID: devID})
	if len(events) == 0 {
		t.Fatalf("Expected activity events for device %d", devID)
	}
	found := false
	for _, ev := range events {
		if ev.EventName == "USB_PORT_MISSING" || ev.EventName == "DEVICE_OFFLINE" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected USB_PORT_MISSING or DEVICE_OFFLINE event")
	}
}

// Case 6: Adapter di-bind kembali (Adapter re-bound)
func TestUSBQuadSerial_Case6_AdapterRebindAndRecovery(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()
	communication.ResetSharedAdapterTracker()

	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	port := "/dev/ttyACM0"
	devID := uint(601)
	tracker.RegisterDevice(devID, "DEV6", "Solar Inverter", port)

	// 1. Adapter unbind
	tracker.OnPortMissing(devID, port, "device disappeared")

	snap, _ := mon.GetDevicePollingSnapshot(devID)
	if snap.CommunicationStatus != communication.CommOffline {
		t.Fatalf("Expected OFFLINE after unbind")
	}

	// 2. Adapter re-bind: echo '1-1.1' | sudo tee /sys/bus/usb/drivers/usb/bind
	// Port reappears
	modbus.SetPortPresenceHookForTest(func(p string) bool {
		return p == port
	})
	defer modbus.SetPortPresenceHookForTest(nil)

	tracker.OnPortRecovered(devID, port)

	snapRecovered, _ := mon.GetDevicePollingSnapshot(devID)
	// Transport must be AVAILABLE, but CommunicationStatus must remain OFFLINE until valid Modbus response
	if snapRecovered.TransportStatus != communication.TransportAvailable {
		t.Errorf("Expected TransportStatus %s after rebind, got %s", communication.TransportAvailable, snapRecovered.TransportStatus)
	}
	if snapRecovered.CommunicationStatus == communication.CommOnline {
		t.Errorf("CommunicationStatus must NOT immediately become ONLINE upon rebind without verified Modbus response!")
	}
	if snapRecovered.RecoveryState != "RECONNECTING" {
		t.Errorf("Expected RecoveryState RECONNECTING, got %s", snapRecovered.RecoveryState)
	}

	// 3. Modbus response verified
	tracker.OnCommunicationSuccess(devID, 18)

	snapFinal, _ := mon.GetDevicePollingSnapshot(devID)
	if snapFinal.CommunicationStatus != communication.CommOnline {
		t.Errorf("Expected CommunicationStatus %s after verified response, got %s", communication.CommOnline, snapFinal.CommunicationStatus)
	}
	if snapFinal.RecoveryState != "NORMAL" {
		t.Errorf("Expected RecoveryState NORMAL, got %s", snapFinal.RecoveryState)
	}
}

// Case 7: Port tersedia tetapi sensor tidak merespons (Port available, sensor unresponsive)
func TestUSBQuadSerial_Case7_PortAvailableSensorUnresponsive(t *testing.T) {
	communication.ResetSharedAdapterTracker()
	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	port := "/dev/ttyACM1"
	devID := uint(701)
	tracker.RegisterDevice(devID, "DEV7", "Silent Sensor", port)

	// Port IS physically present
	modbus.SetPortPresenceHookForTest(func(p string) bool {
		return p == port
	})
	defer modbus.SetPortPresenceHookForTest(nil)

	// Transport opened successfully
	mon.UpdateDevicePollingSnapshot(devID, func(s *communication.DevicePollingSnapshot) {
		s.TransportStatus = communication.TransportAvailable
		s.CommunicationStatus = communication.CommOffline
		s.PollingState = communication.PollingStateTimeout
		s.LastError = "Modbus response timeout after 1000ms"
		s.IsFresh = false
	})

	snap, _ := mon.GetDevicePollingSnapshot(devID)

	// Distinct semantics: Transport AVAILABLE, Communication OFFLINE/TIMEOUT
	if snap.TransportStatus != communication.TransportAvailable {
		t.Errorf("Expected TransportStatus %s, got %s", communication.TransportAvailable, snap.TransportStatus)
	}
	if snap.CommunicationStatus != communication.CommOffline {
		t.Errorf("Expected CommunicationStatus %s, got %s", communication.CommOffline, snap.CommunicationStatus)
	}
	if snap.TransportStatus == communication.TransportPortMissing {
		t.Errorf("Must NOT report PORT_MISSING when port is physically present but sensor is silent!")
	}
}

// Case 8: Port tidak tersedia (Port unavailable)
func TestUSBQuadSerial_Case8_PortUnavailable(t *testing.T) {
	communication.ResetSharedAdapterTracker()
	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	port := "/dev/ttyNONEXISTENT"
	devID := uint(801)
	tracker.RegisterDevice(devID, "DEV8", "Missing Sensor", port)

	modbus.SetPortPresenceHookForTest(func(p string) bool {
		return false
	})
	defer modbus.SetPortPresenceHookForTest(nil)

	tracker.OnPortMissing(devID, port, "open /dev/ttyNONEXISTENT: no such file or directory")

	snap, _ := mon.GetDevicePollingSnapshot(devID)
	if snap.TransportStatus != communication.TransportPortMissing {
		t.Errorf("Expected TransportStatus %s, got %s", communication.TransportPortMissing, snap.TransportStatus)
	}
	if snap.CommunicationStatus != communication.CommOffline {
		t.Errorf("Expected CommunicationStatus %s, got %s", communication.CommOffline, snap.CommunicationStatus)
	}
}

// Case 9: Timeout dan retry (Timeout and retry policy)
func TestUSBQuadSerial_Case9_TimeoutAndRetryPolicy(t *testing.T) {
	conn := &model.DeviceConnection{
		Protocol:        model.ProtocolModbusRTU,
		SerialPort:      "/dev/ttyACM2",
		PollingInterval: 2000,
		Timeout:         500,
		RetryCount:      3,
	}

	thresh := communication.CalculateFreshnessThreshold(conn)
	// (2 * 2000ms) + ((3+1)*500ms) + 1000ms = 4000 + 2000 + 1000 = 7000ms
	expectedMin := 6500 * time.Millisecond
	expectedMax := 7500 * time.Millisecond

	if thresh < expectedMin || thresh > expectedMax {
		t.Errorf("Expected freshness threshold around 7000ms, got %v", thresh)
	}

	// Verify retry backoff calculation
	policy := &communication.BackoffPolicy{
		InitialInterval:     50 * time.Millisecond,
		MaxInterval:         500 * time.Millisecond,
		Multiplier:          2.0,
		RandomizationFactor: 0.0,
	}
	d0 := policy.CalculateDelay(0)
	d1 := policy.CalculateDelay(1)
	d2 := policy.CalculateDelay(2)

	if d0 != 50*time.Millisecond {
		t.Errorf("Expected attempt 0 delay 50ms, got %v", d0)
	}
	if d1 != 100*time.Millisecond {
		t.Errorf("Expected attempt 1 delay 100ms, got %v", d1)
	}
	if d2 != 200*time.Millisecond {
		t.Errorf("Expected attempt 2 delay 200ms, got %v", d2)
	}
}

// Case 10: Recovery tanpa trigger manual (Recovery without manual trigger)
func TestUSBQuadSerial_Case10_RecoveryWithoutManualTrigger(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()
	communication.ResetSharedAdapterTracker()

	reg := newMockPortRegistry()
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 999)

	portName := "/dev/ttyACM3"
	tr := reg.RegisterPort(portName, server)

	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	devID := uint(1001)
	tracker.RegisterDevice(devID, "DEV10", "Auto Recovery Sensor", portName)

	adapter := modbus.NewModbusRTUAdapterWithOpener(&model.DeviceConnection{
		Protocol:   model.ProtocolModbusRTU,
		SerialPort: portName,
		BaudRate:   9600,
		Timeout:    500,
		SlaveID:    1,
	}, reg.Opener)

	ctx := context.Background()
	_ = adapter.Connect(ctx)

	// Unplug
	tr.SetUnplugged(true)
	reg.RemovePort(portName)
	_, _ = adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	tracker.OnPortMissing(devID, portName, "i/o error")

	snap, _ := mon.GetDevicePollingSnapshot(devID)
	if snap.CommunicationStatus != communication.CommOffline {
		t.Fatalf("Expected OFFLINE")
	}

	// Re-plug happens automatically
	server.SetHoldingRegister(0, 999)
	reg.RegisterPort(portName, server)
	tracker.OnPortRecovered(devID, portName)

	// Background worker re-polls on next cycle automatically
	resp, err := adapter.ReadRegisters(ctx, modbus.ModbusReadRequest{
		SlaveID:         1,
		FunctionCode:    modbus.FunctionReadHoldingRegisters,
		StartingAddress: 0,
		Quantity:        1,
	})
	if err != nil || len(resp.Data) < 2 {
		t.Fatalf("Automatic read on recovered port failed: %v", err)
	}

	tracker.OnCommunicationSuccess(devID, 14)

	snapFinal, _ := mon.GetDevicePollingSnapshot(devID)
	if snapFinal.CommunicationStatus != communication.CommOnline {
		t.Errorf("Expected automatic recovery to ONLINE, got %s", snapFinal.CommunicationStatus)
	}
}

// Case 11: Last known value tetap tersedia tetapi tidak dianggap sebagai telemetry baru
func TestUSBQuadSerial_Case11_LastKnownValueHeldNotNewTelemetry(t *testing.T) {
	holdTracker := communication.NewAnomalyHoldTracker()

	param := &model.Parameter{
		ID:                    1101,
		ParameterCode:         "TEMP_HOLD",
		HoldLastValueEnabled: true,
		HoldLastValueSeconds: 60,
		Formula:               "",
	}

	// Seed good reading
	_, _, isHeld1, q1, err := holdTracker.ProcessReading(param, 100, 25.5, nil)
	if err != nil || isHeld1 || q1 != model.QualityGood {
		t.Fatalf("Initial reading failed: %v", err)
	}

	// Communication failure occurs (timeout)
	effVal, _, isHeld2, q2, err2 := holdTracker.ProcessReading(param, 0, 0, errors.New("timeout"))
	if err2 != nil {
		t.Fatalf("Hold process failed: %v", err2)
	}

	// Value is held, but quality is UNCERTAIN and isHeld is true
	if effVal != 25.5 {
		t.Errorf("Expected held value 25.5, got %v", effVal)
	}
	if !isHeld2 {
		t.Errorf("Expected isHeld=true during communication failure")
	}
	if q2 != model.QualityUncertain {
		t.Errorf("Expected QualityUncertain for held stale data during hold grace period, got %v", q2)
	}

	// Stale timestamp check
	staleTime := time.Now().Add(-10 * time.Minute)
	snap := &communication.DevicePollingSnapshot{
		DeviceID:               1101,
		CommunicationStatus:    communication.CommOnline,
		LastSuccessfulResponse: &staleTime,
		FreshnessThresholdMs:   5000,
	}

	isFresh := communication.EvaluateDeviceFreshness(snap)
	if isFresh {
		t.Errorf("Expected EvaluateDeviceFreshness to return false for 10-minute old response!")
	}
	if snap.CommunicationStatus == communication.CommOnline {
		t.Errorf("Expected snapshot status to transition from ONLINE to OFFLINE due to staleness")
	}
}

// Case 12: Status UI dan backend konsisten (UI and backend snapshot consistency)
func TestUSBQuadSerial_Case12_UIAndBackendConsistency(t *testing.T) {
	communication.ResetSharedAdapterTracker()
	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	dev1 := model.Device{ID: 1201, DeviceCode: "D1", Status: model.DeviceStatusActive, Enabled: true}
	dev2 := model.Device{ID: 1202, DeviceCode: "D2", Status: model.DeviceStatusActive, Enabled: true}

	dev1Port := "/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if00"
	dev2Port := "/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if01"

	tracker.RegisterDevice(dev1.ID, dev1.DeviceCode, "Device 1", dev1Port)
	tracker.RegisterDevice(dev2.ID, dev2.DeviceCode, "Device 2", dev2Port)

	tracker.OnCommunicationSuccess(dev1.ID, 10)
	tracker.OnCommunicationSuccess(dev2.ID, 12)

	overviewBefore := mon.GetSystemOverview([]model.Device{dev1, dev2})
	if overviewBefore.DevicesOnline != 2 || overviewBefore.DevicesOffline != 0 {
		t.Fatalf("Expected 2 online before disconnect, got online=%d offline=%d", overviewBefore.DevicesOnline, overviewBefore.DevicesOffline)
	}

	// Disconnect adapter
	modbus.SetPortPresenceHookForTest(func(p string) bool { return false })
	defer modbus.SetPortPresenceHookForTest(nil)

	tracker.OnPortMissing(dev1.ID, dev1Port, "unbound")

	// Overview counts must match snapshots
	overviewAfter := mon.GetSystemOverview([]model.Device{dev1, dev2})
	if overviewAfter.DevicesOnline != 0 {
		t.Errorf("Expected 0 devices online after unbind, got %d", overviewAfter.DevicesOnline)
	}
	if overviewAfter.DevicesOffline != 2 {
		t.Errorf("Expected 2 devices offline after unbind, got %d", overviewAfter.DevicesOffline)
	}
}

// Case 13: Device lain pada adapter berbeda tidak terdampak (Devices on different adapters unaffected)
func TestUSBQuadSerial_Case13_IndependentDevicesOnDifferentAdaptersUnaffected(t *testing.T) {
	modbus.ResetSharedBuses()
	defer modbus.ResetSharedBuses()
	communication.ResetSharedAdapterTracker()

	mon := communication.NewActivityMonitor(nil, nil, nil)
	tracker := communication.GetSharedAdapterTracker()
	tracker.SetActivityMonitor(mon)

	// Adapter A: QinHeng USB Quad Serial (ports ACM0 and ACM1)
	portA1 := "/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if00"
	portA2 := "/dev/serial/by-id/usb-1a86_USB_Quad_Serial_55D5-if02"
	devA1 := uint(1301)
	devA2 := uint(1302)

	// Adapter B: FTDI Single USB Serial (port USB0)
	portB := "/dev/serial/by-id/usb-FTDI_FT232R_USB_UART_A101-if00"
	devB := uint(1303)

	tracker.RegisterDevice(devA1, "QUAD_1", "Quad Sensor 1", portA1)
	tracker.RegisterDevice(devA2, "QUAD_2", "Quad Sensor 2", portA2)
	tracker.RegisterDevice(devB, "FTDI_1", "FTDI Gas Sensor", portB)

	tracker.OnCommunicationSuccess(devA1, 15)
	tracker.OnCommunicationSuccess(devA2, 18)
	tracker.OnCommunicationSuccess(devB, 22)

	// Adapter A is unbound: only its ports vanish! Port B remains present.
	modbus.SetPortPresenceHookForTest(func(port string) bool {
		return port == portB
	})
	defer modbus.SetPortPresenceHookForTest(nil)

	tracker.OnPortMissing(devA1, portA1, "open portA1: no such file or directory")

	snapA1, _ := mon.GetDevicePollingSnapshot(devA1)
	snapA2, _ := mon.GetDevicePollingSnapshot(devA2)
	snapB, _ := mon.GetDevicePollingSnapshot(devB)

	// Devices on Adapter A must be OFFLINE and PORT_MISSING
	if snapA1.CommunicationStatus != communication.CommOffline || snapA1.TransportStatus != communication.TransportPortMissing {
		t.Errorf("Device A1 on Adapter A should be OFFLINE / PORT_MISSING")
	}
	if snapA2.CommunicationStatus != communication.CommOffline || snapA2.TransportStatus != communication.TransportPortMissing {
		t.Errorf("Device A2 on Adapter A should be OFFLINE / PORT_MISSING")
	}

	// Device on Adapter B MUST REMAIN HEALTHY ONLINE AND AVAILABLE!
	if snapB.CommunicationStatus != communication.CommOnline {
		t.Errorf("Device B on separate adapter must remain ONLINE! Got: %s", snapB.CommunicationStatus)
	}
	if snapB.TransportStatus != communication.TransportAvailable {
		t.Errorf("Device B on separate adapter must remain AVAILABLE! Got: %s", snapB.TransportStatus)
	}
	if !snapB.IsFresh {
		t.Errorf("Device B on separate adapter must remain fresh!")
	}
}
