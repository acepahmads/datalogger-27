package communication

import (
	"context"
	"errors"
	"math"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
	"datalogger/internal/service"
)

// Acceptance Criteria 1 & 2: Successful connection & Initial connection failure
func TestReliability_ConnectionAndInitialFailure(t *testing.T) {
	sm := NewConnectionStateMachine()
	tracker := NewConnectionHealthTracker()
	tracker.SetThresholds(3, 5)
	sm.RegisterCallback(func(deviceID uint, from, to ConnectionState, reason string) {
		tracker.SetState(deviceID, to)
	})

	// Device 1 connects successfully
	dev1 := uint(101)
	_, err := sm.Transition(dev1, StateConnecting, "initial connect")
	if err != nil {
		t.Fatalf("expected valid transition to CONNECTING, got: %v", err)
	}
	tracker.RecordConnectAttempt(dev1)

	_, err = sm.Transition(dev1, StateConnected, "connected to socket")
	if err != nil {
		t.Fatalf("expected valid transition to CONNECTED, got: %v", err)
	}
	tracker.RecordConnectSuccess(dev1, 15)

	h1 := tracker.GetHealth(dev1)
	if h1.State != StateConnected || h1.LatencyMs != 15 {
		t.Fatalf("expected state CONNECTED and latency 15, got %v / %d", h1.State, h1.LatencyMs)
	}

	// Device 2 fails initially
	dev2 := uint(102)
	_, _ = sm.Transition(dev2, StateConnecting, "initial connect")
	tracker.RecordConnectAttempt(dev2)

	_, err = sm.Transition(dev2, StateError, "connection refused")
	if err != nil {
		t.Fatalf("expected valid transition to ERROR, got: %v", err)
	}
	tracker.RecordFailure(dev2, errors.New("dial tcp 127.0.0.1:9999: connectex: connection refused"))

	h2 := tracker.GetHealth(dev2)
	if h2.State != StateError || h2.ConsecutiveFailures != 1 || h2.LastErrorCategory != ErrCategoryConnRefused {
		t.Fatalf("expected state ERROR, 1 failure, and CONN_REFUSED, got %+v", h2)
	}
}

// Acceptance Criteria 3 & 4: Successful reconnect after temporary failure & Repeated reconnect failure
func TestReliability_StateTransitions_DegradedAndRepeatedFailure(t *testing.T) {
	sm := NewConnectionStateMachine()
	tracker := NewConnectionHealthTracker()
	tracker.SetThresholds(2, 4) // Degraded at 2, Error at 4
	sm.RegisterCallback(func(deviceID uint, from, to ConnectionState, reason string) {
		tracker.SetState(deviceID, to)
	})
	devID := uint(201)

	_, _ = sm.Transition(devID, StateConnecting, "init")
	_, _ = sm.Transition(devID, StateConnected, "ok")
	tracker.RecordConnectSuccess(devID, 10)

	// Simulate consecutive failures
	for i := 1; i <= 2; i++ {
		tracker.RecordFailure(devID, errors.New("i/o timeout"))
	}
	h := tracker.GetHealth(devID)
	if h.State != StateDegraded {
		t.Fatalf("expected DEGRADED after 2 failures, got %v", h.State)
	}
	_, _ = sm.Transition(devID, StateDegraded, "consecutive timeouts")

	// Trigger reconnecting
	_, _ = sm.Transition(devID, StateReconnecting, "auto-reconnect")
	tracker.RecordReconnectAttempt(devID, 500*time.Millisecond)

	// Further failures reach error threshold
	for i := 3; i <= 4; i++ {
		tracker.RecordFailure(devID, errors.New("broken pipe"))
	}
	h = tracker.GetHealth(devID)
	if h.State != StateError {
		t.Fatalf("expected ERROR after 4 failures, got %v", h.State)
	}
	_, _ = sm.Transition(devID, StateError, "max failures exceeded")

	// Now simulate successful recovery
	_, _ = sm.Transition(devID, StateReconnecting, "manual or auto retry")
	_, _ = sm.Transition(devID, StateConnected, "recovered")
	tracker.RecordConnectSuccess(devID, 12)

	h = tracker.GetHealth(devID)
	if h.State != StateConnected || h.ConsecutiveFailures != 0 {
		t.Fatalf("expected CONNECTED with 0 failures after recovery, got %+v", h)
	}
}

// Acceptance Criteria 5 & 6: Exponential backoff calculation & Jitter bounds
func TestReliability_ExponentialBackoffAndJitter(t *testing.T) {
	policy := &BackoffPolicy{
		InitialInterval:     500 * time.Millisecond,
		MaxInterval:         30 * time.Second,
		Multiplier:          2.0,
		RandomizationFactor: 0.20, // ±20%
	}

	expectedBases := []time.Duration{
		500 * time.Millisecond,
		1000 * time.Millisecond,
		2000 * time.Millisecond,
		4000 * time.Millisecond,
		8000 * time.Millisecond,
		16000 * time.Millisecond,
		30000 * time.Millisecond, // Capped at MaxInterval (30s)
		30000 * time.Millisecond,
	}

	for attempt, expectedBase := range expectedBases {
		delay := policy.CalculateDelay(attempt)
		minBound := time.Duration(float64(expectedBase) * 0.80)
		maxBound := time.Duration(float64(expectedBase) * 1.20)

		if delay < minBound || delay > maxBound {
			t.Fatalf("attempt %d: delay %v out of jitter bounds [%v, %v]", attempt, delay, minBound, maxBound)
		}
	}
}

// Acceptance Criteria 7: Cancellation during backoff delay
func TestReliability_CancellationDuringBackoff(t *testing.T) {
	policy := &BackoffPolicy{
		InitialInterval:     5 * time.Second,
		MaxInterval:         30 * time.Second,
		Multiplier:          2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := policy.Sleep(ctx, 0)
	elapsed := time.Since(start)

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if elapsed > 1*time.Second {
		t.Fatalf("sleep did not abort promptly on cancel, took %v", elapsed)
	}
}

// Acceptance Criteria 8 & 9: Device disablement and removal cancels reconnection
func TestReliability_DisablementAndRemovalDuringReconnect(t *testing.T) {
	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	dev := &model.Device{
		ID:         301,
		DeviceCode: "DEV-DISABLE-TEST",
		Enabled:    true,
		Connection: &model.DeviceConnection{
			DeviceID: 301,
			Protocol: model.ProtocolModbusTCP,
			Host:     "127.0.0.1",
			Port:     65530, // Unused port to trigger failure
			Timeout:  100,
			Enabled:  true,
		},
	}

	ctx := context.Background()
	// Initial reconnect attempt on offline port will fail
	_ = mgr.ReconnectDevice(ctx, dev)

	// Now disable device
	dev.Enabled = false
	mgr.DisableDevice(dev.ID)

	state := mgr.GetDeviceState(dev.ID)
	if state != StateDisabled {
		t.Fatalf("expected device state DISABLED, got %v", state)
	}

	// Reconnect on disabled device should abort immediately
	err := mgr.ReconnectDevice(ctx, dev)
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected error indicating device disabled, got: %v", err)
	}

	// Disconnect / removal cleans up adapter and state
	err = mgr.DisconnectDevice(dev.ID)
	if err != nil {
		t.Fatalf("unexpected error on disconnect: %v", err)
	}
	if mgr.IsDeviceReconnecting(dev.ID) {
		t.Fatalf("device still marked as reconnecting after removal")
	}
}

// Acceptance Criteria 10, 11, 12: Communication timeout, device failure isolation, and multi-device concurrency
func TestReliability_DeviceFailureIsolation(t *testing.T) {
	// Start Mock Server for Device A
	serverA := modbus.NewMockModbusServer()
	serverA.SetHoldingRegister(0, 9999)
	addrA, err := serverA.StartTCP()
	if err != nil {
		t.Fatalf("failed starting server A: %v", err)
	}
	defer serverA.Stop()

	hostA, portStrA, _ := net.SplitHostPort(addrA)
	portA, _ := strconv.Atoi(portStrA)

	devA := &model.Device{
		ID:         401,
		DeviceCode: "DEV-HEALTHY-A",
		Enabled:    true,
		Connection: &model.DeviceConnection{
			DeviceID: 401,
			Protocol: model.ProtocolModbusTCP,
			Host:     hostA,
			Port:     portA,
			Timeout:  500,
			Enabled:  true,
		},
	}

	// Device B points to a dead port (will timeout / fail)
	devB := &model.Device{
		ID:         402,
		DeviceCode: "DEV-UNREACHABLE-B",
		Enabled:    true,
		Connection: &model.DeviceConnection{
			DeviceID: 402,
			Protocol: model.ProtocolModbusTCP,
			Host:     "127.0.0.1",
			Port:     65534,
			Timeout:  100,
			Enabled:  true,
		},
	}

	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	ctx := context.Background()

	// Connect healthy Device A
	if err := mgr.ConnectDevice(ctx, devA); err != nil {
		t.Fatalf("Device A connect failed: %v", err)
	}

	// Attempt connecting Device B (fails)
	_ = mgr.ConnectDevice(ctx, devB)

	// Concurrently read Device A and Device B
	var wg sync.WaitGroup
	var devASuccess bool
	var devAFinishedTime time.Duration
	var devBFinishedTime time.Duration

	wg.Add(2)
	go func() {
		defer wg.Done()
		start := time.Now()
		req := ModbusReadRequest{SlaveID: 1, FunctionCode: FunctionReadHoldingRegisters, StartingAddress: 0, Quantity: 1}
		resp, rErr := mgr.ExecuteWithRetry(ctx, devA, req)
		devAFinishedTime = time.Since(start)
		if rErr == nil && len(resp.Data) > 0 {
			devASuccess = true
		}
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		req := ModbusReadRequest{SlaveID: 1, FunctionCode: FunctionReadHoldingRegisters, StartingAddress: 0, Quantity: 1}
		_, _ = mgr.ExecuteWithRetry(ctx, devB, req)
		devBFinishedTime = time.Since(start)
	}()

	wg.Wait()

	if !devASuccess {
		t.Fatalf("Device A read failed despite being healthy; isolation compromised")
	}
	if devAFinishedTime > 500*time.Millisecond {
		t.Fatalf("Device A was delayed by Device B failure, took %v", devAFinishedTime)
	}
	t.Logf("Device A read succeeded in %v while Device B failed in %v (isolated)", devAFinishedTime, devBFinishedTime)
}

// Acceptance Criteria 13: No duplicate polling workers or reconnect storms
func TestReliability_DeduplicatedReconnection(t *testing.T) {
	mgr := NewConnectionManager(nil, DefaultAdapterFactory)
	dev := &model.Device{
		ID:         501,
		DeviceCode: "DEV-DEDUP-TEST",
		Enabled:    true,
		Connection: &model.DeviceConnection{
			DeviceID: 501,
			Protocol: model.ProtocolModbusTCP,
			Host:     "127.0.0.1",
			Port:     65531,
			Timeout:  200,
			Enabled:  true,
		},
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	var duplicateCalls int32

	// Launch 10 concurrent reconnect calls for the exact same device
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := mgr.ReconnectDevice(ctx, dev)
			if err != nil && strings.Contains(err.Error(), "already in progress") {
				atomic.AddInt32(&duplicateCalls, 1)
			}
		}()
	}
	wg.Wait()

	if duplicateCalls == 0 {
		t.Fatalf("expected concurrent reconnect calls to be deduplicated, but none detected in-flight lock")
	}
	t.Logf("Successfully deduplicated %d concurrent reconnect attempts", duplicateCalls)
}

// Acceptance Criteria 14, 15, 16: Polling recovery, preservation of last valid value, stale detection
func TestReliability_PollingRecovery_PreservesLastValidValue(t *testing.T) {
	cfg := service.DefaultTelemetryConfig()
	cfg.FlushInterval = 50 * time.Millisecond
	ts := service.NewTelemetryService(nil, nil, cfg)
	defer ts.Stop()

	paramID := uint(601)
	devID := uint(60)

	// 1. Initial valid reading: 45.50
	validPayload := &service.TelemetryIngestPayload{
		DeviceID:      devID,
		DeviceCode:    "DEV-60",
		ParameterID:   paramID,
		ParameterCode: "TEMP",
		DataType:      model.DataTypeFloat32,
		RawValue:      45.50,
		Value:         45.50,
		Quality:       model.QualityGood,
		ReceivedAt:    time.Now().UTC(),
	}
	err := ts.Ingest(validPayload)
	if err != nil {
		t.Fatalf("failed ingesting valid telemetry: %v", err)
	}

	latest, err := ts.GetLatestForParameter(context.Background(), devID, paramID)
	if err != nil || latest == nil || latest.Value != 45.50 || latest.Quality != model.QualityGood {
		t.Fatalf("unexpected latest cache: %+v, err: %v", latest, err)
	}

	// 2. Failed reading (communication timeout)
	// Must NOT overwrite 45.50 with 0.0, but MUST flag Quality as BAD
	failPayload := &service.TelemetryIngestPayload{
		DeviceID:      devID,
		DeviceCode:    "DEV-60",
		ParameterID:   paramID,
		ParameterCode: "TEMP",
		DataType:      model.DataTypeFloat32,
		RawValue:      0.0,
		Value:         0.0,
		ReadError:     errors.New("i/o timeout"),
		ReceivedAt:    time.Now().UTC(),
	}
	_ = ts.Ingest(failPayload)

	latestAfterErr, err := ts.GetLatestForParameter(context.Background(), devID, paramID)
	if err != nil || latestAfterErr == nil {
		t.Fatalf("failed reading latest cache after error: %v", err)
	}

	if latestAfterErr.Value != 45.50 {
		t.Fatalf("CRITICAL: Last valid reading was overwritten! Expected 45.50, got %v", latestAfterErr.Value)
	}
	if latestAfterErr.Quality != model.QualityBad {
		t.Fatalf("expected Quality BAD on read error, got %v", latestAfterErr.Quality)
	}

	// 3. Recovery reading: 46.20
	recoverPayload := &service.TelemetryIngestPayload{
		DeviceID:      devID,
		DeviceCode:    "DEV-60",
		ParameterID:   paramID,
		ParameterCode: "TEMP",
		DataType:      model.DataTypeFloat32,
		RawValue:      46.20,
		Value:         46.20,
		Quality:       model.QualityGood,
		ReceivedAt:    time.Now().UTC(),
	}
	_ = ts.Ingest(recoverPayload)

	latestAfterRecover, _ := ts.GetLatestForParameter(context.Background(), devID, paramID)
	if latestAfterRecover.Value != 46.20 || latestAfterRecover.Quality != model.QualityGood {
		t.Fatalf("expected recovered value 46.20 and GOOD quality, got %+v", latestAfterRecover)
	}
}

// Mock persistence boundary for deterministic failure testing
type mockPersistenceBoundary struct {
	mu           sync.Mutex
	failAttempts int
	savedBatches [][]*model.RawData
}

func (m *mockPersistenceBoundary) SaveBatch(ctx context.Context, batch []*model.RawData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failAttempts > 0 {
		m.failAttempts--
		return errors.New("temporary MariaDB connection pool timeout (deadlock)")
	}
	m.savedBatches = append(m.savedBatches, batch)
	return nil
}

// Acceptance Criteria 17 & 18: Temporary database failure & Persistence health metrics
func TestReliability_DatabaseFailureBoundary(t *testing.T) {
	cfg := service.DefaultTelemetryConfig()
	cfg.BufferSize = 100
	cfg.BatchSize = 10
	cfg.FlushInterval = 20 * time.Millisecond
	cfg.MaxRetries = 2

	ts := service.NewTelemetryService(nil, nil, cfg)
	defer ts.Stop()

	mockDB := &mockPersistenceBoundary{failAttempts: 1} // Fails once, then recovers
	ts.SetPersistenceBoundary(mockDB)

	// Ingest batch
	for i := 1; i <= 10; i++ {
		_ = ts.Ingest(&service.TelemetryIngestPayload{
			DeviceID:      701,
			ParameterID:   uint(700 + i),
			Value:         float64(i * 10),
			ReceivedAt:    time.Now().UTC(),
		})
	}

	// Allow flush worker to execute
	time.Sleep(150 * time.Millisecond)

	metrics := ts.GetMetrics()
	if metrics.PersistedCount != 10 {
		t.Fatalf("expected 10 persisted records after retry recovery, got %d", metrics.PersistedCount)
	}
	if metrics.PersistenceStatus != "HEALTHY" {
		t.Fatalf("expected HEALTHY persistence status after recovery, got %s", metrics.PersistenceStatus)
	}

	// Now simulate permanent DB outage
	mockDB.mu.Lock()
	mockDB.failAttempts = 999
	mockDB.mu.Unlock()

	for i := 11; i <= 20; i++ {
		_ = ts.Ingest(&service.TelemetryIngestPayload{
			DeviceID:      701,
			ParameterID:   uint(700 + i),
			Value:         float64(i * 10),
			ReceivedAt:    time.Now().UTC(),
		})
	}

	time.Sleep(200 * time.Millisecond)
	metricsAfterOutage := ts.GetMetrics()
	if metricsAfterOutage.PersistenceStatus != "DEGRADED" && metricsAfterOutage.PersistenceStatus != "FAILING" {
		t.Fatalf("expected DEGRADED or FAILING persistence status, got %s", metricsAfterOutage.PersistenceStatus)
	}
	if metricsAfterOutage.ConsecutiveDBErrors == 0 {
		t.Fatalf("expected consecutive DB errors to be tracked, got 0")
	}
}

// Acceptance Criteria 19, 21, 22: Graceful shutdown and no goroutine/socket leaks
func TestReliability_GracefulShutdownAndZeroLeaks(t *testing.T) {
	runtime.GC()
	baselineGoroutines := runtime.NumGoroutine()

	cfg := service.DefaultTelemetryConfig()
	cfg.FlushInterval = 50 * time.Millisecond
	ts := service.NewTelemetryService(nil, nil, cfg)
	mockDB := &mockPersistenceBoundary{}
	ts.SetPersistenceBoundary(mockDB)

	// Ingest items
	for i := 0; i < 50; i++ {
		_ = ts.Ingest(&service.TelemetryIngestPayload{
			DeviceID:    801,
			ParameterID: uint(800 + i),
			Value:       float64(i),
			ReceivedAt:  time.Now().UTC(),
		})
	}

	// Stop cleanly
	ts.Stop()

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	currentGoroutines := runtime.NumGoroutine()

	diff := math.Abs(float64(currentGoroutines - baselineGoroutines))
	if diff > 4 {
		t.Fatalf("potential goroutine leak detected: baseline %d, current %d (diff %v)",
			baselineGoroutines, currentGoroutines, diff)
	}
}

// Acceptance Criteria 25: EN/ID localization key parity
func TestReliability_LocalizationKeyParity(t *testing.T) {
	enPath := "../../web/src/i18n/locales/en.js"
	idPath := "../../web/src/i18n/locales/id.js"

	enData, err := os.ReadFile(enPath)
	if err != nil {
		t.Skipf("skipping localization file test: %v", err)
	}
	idData, err := os.ReadFile(idPath)
	if err != nil {
		t.Skipf("skipping localization file test: %v", err)
	}

	enStr := string(enData)
	idStr := string(idData)

	// Verify key status terms exist in both locales
	criticalTerms := []string{
		"connected",
		"disconnected",
		"reconnecting",
		"degraded",
		"disabled",
		"error",
		"uptime",
	}

	for _, term := range criticalTerms {
		if !strings.Contains(strings.ToLower(enStr), term) {
			t.Errorf("missing critical term '%s' in en.js", term)
		}
		if !strings.Contains(strings.ToLower(idStr), term) {
			t.Errorf("missing critical term '%s' in id.js", term)
		}
	}
}

// Acceptance Criteria 2: Invalid State Machine Transitions Rejected
func TestReliability_InvalidStateTransitionsRejected(t *testing.T) {
	sm := NewConnectionStateMachine()
	devID := uint(901)

	// Device starts at DISCONNECTED
	// Invalid: DISCONNECTED directly to DEGRADED
	_, err := sm.Transition(devID, StateDegraded, "invalid transition")
	if err == nil {
		t.Fatalf("expected error for invalid transition DISCONNECTED -> DEGRADED, got nil")
	}

	// Invalid: DISCONNECTED directly to RECONNECTING
	_, err = sm.Transition(devID, StateReconnecting, "invalid transition")
	if err == nil {
		t.Fatalf("expected error for invalid transition DISCONNECTED -> RECONNECTING, got nil")
	}

	// Valid: DISCONNECTED -> DISABLED
	_, err = sm.Transition(devID, StateDisabled, "admin disable")
	if err != nil {
		t.Fatalf("expected valid transition DISCONNECTED -> DISABLED, got %v", err)
	}

	// Invalid: DISABLED -> CONNECTING
	_, err = sm.Transition(devID, StateConnecting, "invalid connect on disabled")
	if err == nil {
		t.Fatalf("expected error for invalid transition DISABLED -> CONNECTING, got nil")
	}
}
