package communication

import (
	"strings"
	"sync"
	"time"
)

// Standardized error categories
const (
	ErrCategoryTimeout        = "TIMEOUT"
	ErrCategoryConnRefused    = "CONNECTION_REFUSED"
	ErrCategoryProtocol       = "PROTOCOL_EXCEPTION"
	ErrCategoryConfig         = "CONFIGURATION_ERROR"
	ErrCategoryCommFailure    = "COMMUNICATION_FAILURE"
)

// CategorizeCommunicationError classifies error strings into structured categories
func CategorizeCommunicationError(err error) string {
	if err == nil {
		return ""
	}
	s := strings.ToLower(err.Error())
	switch {
	case strings.Contains(s, "timeout") || strings.Contains(s, "deadline exceeded") || strings.Contains(s, "i/o timeout"):
		return ErrCategoryTimeout
	case strings.Contains(s, "refused") || strings.Contains(s, "reset by peer") || strings.Contains(s, "broken pipe") || strings.Contains(s, "offline") || strings.Contains(s, "dial failed"):
		return ErrCategoryConnRefused
	case strings.Contains(s, "illegal") || strings.Contains(s, "exception") || strings.Contains(s, "crc") || strings.Contains(s, "checksum") || strings.Contains(s, "framing"):
		return ErrCategoryProtocol
	case strings.Contains(s, "invalid port") || strings.Contains(s, "not found") || strings.Contains(s, "no connection"):
		return ErrCategoryConfig
	default:
		return ErrCategoryCommFailure
	}
}

// ConnectionHealth captures runtime observability metrics and timestamps for a device
type ConnectionHealth struct {
	DeviceID                  uint            `json:"device_id"`
	State                     ConnectionState `json:"state"`
	ConnectedSince            *time.Time      `json:"connected_since,omitempty"`
	LastConnectionAttempt     *time.Time      `json:"last_connection_attempt,omitempty"`
	LastSuccessfulConnection  *time.Time      `json:"last_successful_connection,omitempty"`
	LastCommSuccess           *time.Time      `json:"last_comm_success,omitempty"`
	LastCommFailure           *time.Time      `json:"last_comm_failure,omitempty"`
	ConsecutiveFailures       int             `json:"consecutive_failures"`
	ConsecutiveSuccesses      int             `json:"consecutive_successes"`
	TotalReconnectAttempts    uint64          `json:"total_reconnect_attempts"`
	TotalSuccessfulReconnects uint64          `json:"total_successful_reconnects"`
	NextReconnectTime         *time.Time      `json:"next_reconnect_time,omitempty"`
	LastErrorCategory         string          `json:"last_error_category,omitempty"`
	LastError                 string          `json:"last_error,omitempty"`
	LastErrorTime             *time.Time      `json:"last_error_time,omitempty"`
	UptimeSeconds             uint64          `json:"uptime_seconds"`
	LatencyMs                 int             `json:"latency_ms"`
	CurrentBackoffDelayMs     int64           `json:"current_backoff_delay_ms"`
	DegradedThreshold         int             `json:"degraded_threshold"`
	ErrorThreshold            int             `json:"error_threshold"`
}

// ConnectionHealthTracker tracks and calculates live health per device
type ConnectionHealthTracker struct {
	mu                sync.RWMutex
	records           map[uint]*ConnectionHealth
	degradedThreshold int
	errorThreshold    int
}

// NewConnectionHealthTracker constructs the tracker with standard thresholds
func NewConnectionHealthTracker() *ConnectionHealthTracker {
	return &ConnectionHealthTracker{
		records:           make(map[uint]*ConnectionHealth),
		degradedThreshold: 3,  // 3 consecutive failures -> DEGRADED
		errorThreshold:    10, // 10 consecutive failures -> ERROR
	}
}

// SetThresholds configures consecutive failure thresholds for DEGRADED and ERROR
func (t *ConnectionHealthTracker) SetThresholds(degraded, errorThreshold int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if degraded > 0 {
		t.degradedThreshold = degraded
	}
	if errorThreshold > 0 {
		t.errorThreshold = errorThreshold
	}
}

func (t *ConnectionHealthTracker) getOrCreate(deviceID uint) *ConnectionHealth {
	h, exists := t.records[deviceID]
	if !exists {
		h = &ConnectionHealth{
			DeviceID:          deviceID,
			State:             StateDisconnected,
			DegradedThreshold: t.degradedThreshold,
			ErrorThreshold:    t.errorThreshold,
		}
		t.records[deviceID] = h
	}
	return h
}

// RecordConnectionAttempt records a connection start
func (t *ConnectionHealthTracker) RecordConnectionAttempt(deviceID uint) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	h := t.getOrCreate(deviceID)
	h.LastConnectionAttempt = &now
}

// RecordConnectionSuccess records connection establishment
func (t *ConnectionHealthTracker) RecordConnectionSuccess(deviceID uint) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	h := t.getOrCreate(deviceID)
	h.LastSuccessfulConnection = &now
	h.ConnectedSince = &now
	h.State = StateConnected
	h.LastError = ""
	h.LastErrorCategory = ""
	h.CurrentBackoffDelayMs = 0
	h.NextReconnectTime = nil
}

// RecordCommunicationSuccess records a successful telemetry read
func (t *ConnectionHealthTracker) RecordCommunicationSuccess(deviceID uint, latencyMs int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	h := t.getOrCreate(deviceID)
	h.LastCommSuccess = &now
	h.ConsecutiveSuccesses++
	h.ConsecutiveFailures = 0
	h.LatencyMs = latencyMs
	h.CurrentBackoffDelayMs = 0
	h.NextReconnectTime = nil
	if h.ConnectedSince != nil {
		h.UptimeSeconds = uint64(now.Sub(*h.ConnectedSince).Seconds())
	}
}

// RecordCommunicationFailure records a communication error
func (t *ConnectionHealthTracker) RecordCommunicationFailure(deviceID uint, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	h := t.getOrCreate(deviceID)
	h.LastCommFailure = &now
	h.LastErrorTime = &now
	h.ConsecutiveFailures++
	h.ConsecutiveSuccesses = 0
	h.LastErrorCategory = CategorizeCommunicationError(err)
	if err != nil {
		h.LastError = err.Error()
	}
	if h.State != StateDisabled {
		if h.ConsecutiveFailures >= t.errorThreshold {
			h.State = StateError
		} else if h.ConsecutiveFailures >= t.degradedThreshold {
			h.State = StateDegraded
		}
	}
}

// RecordFailure is an alias for RecordCommunicationFailure
func (t *ConnectionHealthTracker) RecordFailure(deviceID uint, err error) {
	t.RecordCommunicationFailure(deviceID, err)
}

// RecordConnectAttempt is an alias for RecordConnectionAttempt
func (t *ConnectionHealthTracker) RecordConnectAttempt(deviceID uint) {
	t.RecordConnectionAttempt(deviceID)
}

// RecordConnectSuccess records connection and initial latency
func (t *ConnectionHealthTracker) RecordConnectSuccess(deviceID uint, latencyMs int) {
	t.RecordConnectionSuccess(deviceID)
	t.RecordCommunicationSuccess(deviceID, latencyMs)
}

// RecordReconnectAttempt records an in-progress reconnection cycle
func (t *ConnectionHealthTracker) RecordReconnectAttempt(deviceID uint, nextRetryDelay time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	h := t.getOrCreate(deviceID)
	h.TotalReconnectAttempts++
	h.LastConnectionAttempt = &now
	h.CurrentBackoffDelayMs = nextRetryDelay.Milliseconds()
	next := now.Add(nextRetryDelay)
	h.NextReconnectTime = &next
}

// RecordReconnectSuccess records that a reconnection cycle succeeded
func (t *ConnectionHealthTracker) RecordReconnectSuccess(deviceID uint) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	h := t.getOrCreate(deviceID)
	h.TotalSuccessfulReconnects++
	h.LastSuccessfulConnection = &now
	h.ConnectedSince = &now
	h.ConsecutiveFailures = 0
	h.CurrentBackoffDelayMs = 0
	h.NextReconnectTime = nil
	h.LastError = ""
	h.LastErrorCategory = ""
}

// SetState updates the tracking state directly
func (t *ConnectionHealthTracker) SetState(deviceID uint, s ConnectionState) {
	t.mu.Lock()
	defer t.mu.Unlock()
	h := t.getOrCreate(deviceID)
	h.State = s
	if s == StateDisconnected || s == StateDisabled {
		h.ConnectedSince = nil
		h.UptimeSeconds = 0
	}
}

// GetHealth returns a snapshot of connection health for a device
func (t *ConnectionHealthTracker) GetHealth(deviceID uint) ConnectionHealth {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if h, exists := t.records[deviceID]; exists {
		res := *h
		if res.ConnectedSince != nil {
			res.UptimeSeconds = uint64(time.Since(*res.ConnectedSince).Seconds())
		}
		return res
	}
	return ConnectionHealth{
		DeviceID:          deviceID,
		State:             StateDisconnected,
		DegradedThreshold: t.degradedThreshold,
		ErrorThreshold:    t.errorThreshold,
	}
}

// GetAllHealth returns snapshots for all tracked devices
func (t *ConnectionHealthTracker) GetAllHealth() map[uint]ConnectionHealth {
	t.mu.RLock()
	defer t.mu.RUnlock()
	now := time.Now()
	res := make(map[uint]ConnectionHealth, len(t.records))
	for id, h := range t.records {
		cp := *h
		if cp.ConnectedSince != nil {
			cp.UptimeSeconds = uint64(now.Sub(*cp.ConnectedSince).Seconds())
		}
		res[id] = cp
	}
	return res
}

// RemoveDevice clears tracking data for deleted devices
func (t *ConnectionHealthTracker) RemoveDevice(deviceID uint) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.records, deviceID)
}
