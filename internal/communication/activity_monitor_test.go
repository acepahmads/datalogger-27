package communication

import (
	"context"
	"strings"
	"testing"
	"time"

	"datalogger/internal/model"
)

func TestActivityMonitor_RingBufferAndFilters(t *testing.T) {
	mon := NewActivityMonitor(nil, nil, nil)

	// Record events
	for i := 1; i <= 25; i++ {
		sev := SeverityInfo
		if i%5 == 0 {
			sev = SeverityError
		} else if i%3 == 0 {
			sev = SeverityWarning
		}

		cat := CategoryPolling
		if i%2 == 0 {
			cat = CategoryCommunication
		}

		mon.RecordEvent(ActivityEvent{
			Severity:    sev,
			Category:    cat,
			Component:   "TestComponent",
			DeviceID:    uint(i % 3),
			DeviceCode:  "DEV-01",
			EventName:   "TEST_EVENT",
			Description: "Test event execution description",
		})
	}

	// 1. Test total events
	events, total := mon.GetEvents(ActivityFilter{Limit: 50})
	if total != 25 {
		t.Fatalf("expected 25 total events, got %d", total)
	}
	if len(events) != 25 {
		t.Fatalf("expected 25 events returned, got %d", len(events))
	}

	// 2. Test Severity filter
	errEvents, errTotal := mon.GetEvents(ActivityFilter{Severity: SeverityError, Limit: 50})
	if errTotal != 5 {
		t.Errorf("expected 5 ERROR events, got %d", errTotal)
	}
	for _, ev := range errEvents {
		if ev.Severity != SeverityError {
			t.Errorf("expected ERROR severity, got %s", ev.Severity)
		}
	}

	// 3. Test Category filter
	commEvents, commTotal := mon.GetEvents(ActivityFilter{Category: CategoryCommunication, Limit: 50})
	if commTotal != 12 {
		t.Errorf("expected 12 COMMUNICATION events, got %d", commTotal)
	}
	for _, ev := range commEvents {
		if ev.Category != CategoryCommunication {
			t.Errorf("expected COMMUNICATION category, got %s", ev.Category)
		}
	}

	// 4. Test DeviceID filter
	dev1Events, dev1Total := mon.GetEvents(ActivityFilter{DeviceID: 1, Limit: 50})
	if dev1Total <= 0 {
		t.Errorf("expected > 0 events for device 1, got %d", dev1Total)
	}
	for _, ev := range dev1Events {
		if ev.DeviceID != 1 {
			t.Errorf("expected device ID 1, got %d", ev.DeviceID)
		}
	}

	// 5. Test Search filter
	searchEvents, searchTotal := mon.GetEvents(ActivityFilter{Search: "TEST_EVENT", Limit: 50})
	if searchTotal != 25 || len(searchEvents) != 25 {
		t.Errorf("expected 25 events matching 'TEST_EVENT', got %d", searchTotal)
	}
}

func TestActivityMonitor_DevicePollingSnapshot(t *testing.T) {
	mon := NewActivityMonitor(nil, nil, nil)
	deviceID := uint(42)

	// Initially no snapshot exists
	_, exists := mon.GetDevicePollingSnapshot(deviceID)
	if exists {
		t.Fatal("expected snapshot to not exist initially")
	}

	// Update snapshot
	startTime := time.Now()
	mon.UpdateDevicePollingSnapshot(deviceID, func(s *DevicePollingSnapshot) {
		s.DeviceCode = "SENS-01"
		s.DeviceName = "Weather Sensor"
		s.Protocol = "MODBUS_RTU"
		s.CommunicationStatus = "ONLINE"
		s.PollingState = PollingStatePolling
		s.LastPollingStarted = &startTime
		s.TotalParameters = 8
		s.SuccessParameters = 7
		s.FailedParameters = 1
	})

	snap, exists := mon.GetDevicePollingSnapshot(deviceID)
	if !exists {
		t.Fatal("expected snapshot to exist after update")
	}
	if snap.DeviceCode != "SENS-01" || snap.PollingState != PollingStatePolling {
		t.Errorf("unexpected snapshot data: %+v", snap)
	}
	if snap.TotalParameters != 8 || snap.SuccessParameters != 7 || snap.FailedParameters != 1 {
		t.Errorf("unexpected parameter counts: %+v", snap)
	}

	// Verify all snapshots listing
	allSnaps := mon.GetAllPollingSnapshots()
	if len(allSnaps) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(allSnaps))
	}
}

func TestActivityMonitor_SystemOverview(t *testing.T) {
	mon := NewActivityMonitor(nil, nil, nil)

	devices := []model.Device{
		{ID: 1, DeviceCode: "DEV-1", ConnectionStatus: model.DeviceConnOnline},
		{ID: 2, DeviceCode: "DEV-2", ConnectionStatus: model.DeviceConnOffline},
		{ID: 3, DeviceCode: "DEV-3", ConnectionStatus: model.DeviceConnError},
	}

	// Mark device 1 poll started
	mon.MarkPollStarted(1)

	overview := mon.GetSystemOverview(devices)
	if overview.TotalDevices != 3 {
		t.Errorf("expected 3 total devices, got %d", overview.TotalDevices)
	}
	if overview.DevicesOnline != 1 {
		t.Errorf("expected 1 online device, got %d", overview.DevicesOnline)
	}
	if overview.ActivePollsCount != 1 {
		t.Errorf("expected 1 active poll, got %d", overview.ActivePollsCount)
	}
	if overview.OverallStatus != "DEGRADED" {
		t.Errorf("expected DEGRADED overall status when devices are offline, got %s", overview.OverallStatus)
	}

	mon.MarkPollEnded(1)
	overviewAfter := mon.GetSystemOverview(devices)
	if overviewAfter.ActivePollsCount != 0 {
		t.Errorf("expected 0 active polls after ending, got %d", overviewAfter.ActivePollsCount)
	}
}

func TestActivityMonitor_BackgroundJobs(t *testing.T) {
	mon := NewActivityMonitor(nil, nil, nil)

	jobs := mon.GetAllBackgroundJobs()
	if len(jobs) < 5 {
		t.Fatalf("expected at least 5 default background jobs, got %d", len(jobs))
	}

	// Test update
	mon.UpdateBackgroundJob("polling_scheduler", "COMMUNICATION", "RUNNING", 15, "Executed normally", "", nil, "Detailed state")
	jobsUpdated := mon.GetAllBackgroundJobs()
	found := false
	for _, j := range jobsUpdated {
		if j.JobName == "polling_scheduler" {
			found = true
			if j.LastDurationMs != 15 || j.LastResult != "Executed normally" {
				t.Errorf("unexpected job details: %+v", j)
			}
		}
	}
	if !found {
		t.Error("polling_scheduler job not found in background jobs list")
	}
}

func TestActivityMonitor_Export(t *testing.T) {
	mon := NewActivityMonitor(nil, nil, nil)
	mon.RecordEvent(ActivityEvent{
		Severity:    SeverityInfo,
		Category:    CategorySystem,
		Component:   "Kernel",
		DeviceID:    1,
		DeviceCode:  "DEV-01",
		EventName:   "BOOT_COMPLETE",
		Description: "System activity monitoring engine online",
		DurationMs:  12,
	})

	// CSV Export
	csvData, mime, err := mon.ExportEvents(ActivityFilter{}, "csv")
	if err != nil {
		t.Fatalf("unexpected CSV export error: %v", err)
	}
	if mime != "text/csv" {
		t.Errorf("expected text/csv, got %s", mime)
	}
	csvStr := string(csvData)
	if !strings.Contains(csvStr, "BOOT_COMPLETE") || !strings.Contains(csvStr, "DEV-01") {
		t.Errorf("CSV export missing expected contents: %s", csvStr)
	}

	// JSON Export
	jsonData, jMime, err := mon.ExportEvents(ActivityFilter{}, "json")
	if err != nil {
		t.Fatalf("unexpected JSON export error: %v", err)
	}
	if jMime != "application/json" {
		t.Errorf("expected application/json, got %s", jMime)
	}
	if !strings.Contains(string(jsonData), "BOOT_COMPLETE") {
		t.Errorf("JSON export missing expected contents: %s", string(jsonData))
	}
}

func TestConnectionManager_ReconnectDoesNotPrematurelySetOnline(t *testing.T) {
	// Verify that ReconnectDevice reopens transport but DOES NOT set device to ONLINE without Modbus proof
	dev := &model.Device{
		ID:         99,
		DeviceCode: "TEST-RTU",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
		Connection: &model.DeviceConnection{
			Protocol:        model.ProtocolModbusRTU,
			SerialPort:      "COM999",
			PollingInterval: 1000,
			Timeout:         50,
			RetryCount:      1,
		},
	}

	mockFactory := func(cfg *model.DeviceConnection) (ProtocolAdapter, error) {
		return &mockTestAdapter{connected: false}, nil
	}

	cm := NewConnectionManager(nil, mockFactory)
	mon := NewActivityMonitor(nil, nil, nil)
	cm.SetActivityMonitor(mon)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := cm.ReconnectDevice(ctx, dev)
	if err != nil {
		t.Fatalf("expected mock reconnect to succeed, got %v", err)
	}

	// State machine should be Connected (transport ready)
	if cm.GetStateMachine().GetState(dev.ID) != StateConnected {
		t.Errorf("expected StateConnected, got %v", cm.GetStateMachine().GetState(dev.ID))
	}

	// But in ActivityMonitor snapshot, it is waiting for verified Modbus response (NOT forced ONLINE)
	snap, exists := mon.GetDevicePollingSnapshot(dev.ID)
	if !exists {
		t.Fatal("expected snapshot after reconnect")
	}
	if snap.PollingState == PollingStateSuccess {
		t.Errorf("expected PollingState waiting or connecting, not prematurely SUCCESS: %v", snap.PollingState)
	}
}

type mockTestAdapter struct {
	connected bool
}

func (m *mockTestAdapter) Connect(ctx context.Context) error {
	m.connected = true
	return nil
}
func (m *mockTestAdapter) Disconnect() error {
	m.connected = false
	return nil
}
func (m *mockTestAdapter) IsConnected() bool {
	return m.connected
}
func (m *mockTestAdapter) ReadRegisters(ctx context.Context, req ModbusReadRequest) (*ModbusReadResponse, error) {
	return &ModbusReadResponse{Data: []byte{0x00, 0x01}, ResponseTime: 10 * time.Millisecond}, nil
}
func (m *mockTestAdapter) WriteRegisters(ctx context.Context, req ModbusWriteRequest) (*ModbusWriteResponse, error) {
	return &ModbusWriteResponse{SlaveID: req.SlaveID, FunctionCode: req.FunctionCode, Address: req.Address, Value: req.Value, ResponseTime: 10 * time.Millisecond, CRCPassed: true}, nil
}
func (m *mockTestAdapter) HealthCheck(ctx context.Context) error {
	return nil
}
func (m *mockTestAdapter) GetStatus() AdapterStatus {
	return AdapterStatus{State: StateConnected}
}
func (m *mockTestAdapter) GetConfig() *model.DeviceConnection {
	return &model.DeviceConnection{Protocol: model.ProtocolModbusRTU}
}
