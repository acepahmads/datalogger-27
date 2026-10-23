package communication

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"
)

// ActivitySeverity levels
const (
	SeverityDebug    = "DEBUG"
	SeverityInfo     = "INFO"
	SeverityWarning  = "WARNING"
	SeverityError    = "ERROR"
	SeverityCritical = "CRITICAL"
)

// ActivityCategory classifications
const (
	CategoryPolling       = "POLLING"
	CategoryCommunication = "COMMUNICATION"
	CategorySerial        = "SERIAL"
	CategoryModbus        = "MODBUS"
	CategoryTelemetry     = "TELEMETRY"
	CategoryAlarm         = "ALARM"
	CategoryScheduler     = "SCHEDULER"
	CategorySystem        = "SYSTEM"
	CategoryDatabase      = "DATABASE"
	CategorySecurity      = "SECURITY"
)

// PollingState represents the immediate operational phase of a device's polling loop
type PollingState string

const (
	PollingStateWaiting      PollingState = "WAITING"
	PollingStatePolling      PollingState = "POLLING"
	PollingStateSuccess      PollingState = "SUCCESS"
	PollingStateTimeout      PollingState = "TIMEOUT"
	PollingStateRetrying     PollingState = "RETRYING"
	PollingStateReconnecting PollingState = "RECONNECTING"
	PollingStateRecovered    PollingState = "RECOVERED"
	PollingStateStopped      PollingState = "STOPPED"
	PollingStateUnknown      PollingState = "UNKNOWN"
)

// ActivityEvent represents a structured, high-visibility event in the system activity monitor
type ActivityEvent struct {
	ID            int64                  `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	Severity      string                 `json:"severity"` // DEBUG, INFO, WARNING, ERROR, CRITICAL
	Category      string                 `json:"category"` // POLLING, COMMUNICATION, SERIAL, MODBUS, etc.
	Component     string                 `json:"component"`
	DeviceID      uint                   `json:"device_id,omitempty"`
	DeviceCode    string                 `json:"device_code,omitempty"`
	DeviceName    string                 `json:"device_name,omitempty"`
	EventName     string                 `json:"event_name"`
	Description   string                 `json:"description"`
	DurationMs    int                    `json:"duration_ms,omitempty"`
	ErrorDetail   string                 `json:"error_detail,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DevicePollingSnapshot captures real-time communication & polling observability for a device
type DevicePollingSnapshot struct {
	DeviceID                uint         `json:"device_id"`
	DeviceCode              string       `json:"device_code"`
	DeviceName              string       `json:"device_name"`
	Protocol                string       `json:"protocol"`
	Transport               string       `json:"transport"`
	CommunicationStatus     string       `json:"communication_status"` // ONLINE, OFFLINE, DEGRADED, RECONNECTING, ERROR
	PollingState            PollingState `json:"polling_state"`        // WAITING, POLLING, SUCCESS, TIMEOUT, RETRYING, etc.
	LastPollingStarted      *time.Time   `json:"last_polling_started,omitempty"`
	LastPollingCompleted    *time.Time   `json:"last_polling_completed,omitempty"`
	LastSuccessfulResponse  *time.Time   `json:"last_successful_response,omitempty"`
	LastSuccessfulTelemetry *time.Time   `json:"last_successful_telemetry,omitempty"`
	PollingDurationMs       int          `json:"polling_duration_ms"`
	LastLatencyMs           int          `json:"last_latency_ms"`
	ConsecutiveFailures     int          `json:"consecutive_failures"`
	RetryCount              int          `json:"retry_count"`
	NextScheduledPolling    *time.Time   `json:"next_scheduled_polling,omitempty"`
	LastError               string       `json:"last_error,omitempty"`
	RecoveryState           string       `json:"recovery_state"` // NORMAL, BACKOFF, RETRYING, RECONNECTING
	TotalParameters         int          `json:"total_parameters"`
	SuccessParameters       int          `json:"success_parameters"`
	FailedParameters        int          `json:"failed_parameters"`
	PendingParameters       int          `json:"pending_parameters"`
}

// BackgroundJobSnapshot captures the execution status of system background routines
type BackgroundJobSnapshot struct {
	JobName        string     `json:"job_name"`
	Category       string     `json:"category"`
	Status         string     `json:"status"` // RUNNING, IDLE, DEGRADED, STOPPED, NOT_AVAILABLE
	LastStarted    *time.Time `json:"last_started,omitempty"`
	LastCompleted  *time.Time `json:"last_completed,omitempty"`
	LastDurationMs int        `json:"last_duration_ms,omitempty"`
	LastResult     string     `json:"last_result,omitempty"`
	LastError      string     `json:"last_error,omitempty"`
	NextRun        *time.Time `json:"next_run,omitempty"`
	Details        string     `json:"details,omitempty"`
}

// SystemOverviewSnapshot presents executive health and live metric counters
type SystemOverviewSnapshot struct {
	OverallStatus         string     `json:"overall_status"` // RUNNING, IDLE, DEGRADED, STOPPED, UNKNOWN
	BackendServiceStatus  string     `json:"backend_service_status"`
	PollingEngineStatus   string     `json:"polling_engine_status"`
	DevicesOnline         int        `json:"devices_online"`
	DevicesOffline        int        `json:"devices_offline"`
	DevicesReconnecting   int        `json:"devices_reconnecting"`
	TotalDevices          int        `json:"total_devices"`
	ActivePollsCount      int        `json:"active_polls_count"`
	WaitingOrTimeoutCount int        `json:"waiting_or_timeout_count"`
	ActiveErrorsCount     int        `json:"active_errors_count"`
	LastActivityTime      *time.Time `json:"last_activity_time,omitempty"`
	ServerTime            time.Time  `json:"server_time"`
	UptimeSeconds         int64      `json:"uptime_seconds"`
}

// ActivityFilter options for querying the event feed
type ActivityFilter struct {
	Search    string
	Severity  string
	Category  string
	DeviceID  uint
	StartTime *time.Time
	EndTime   *time.Time
	SinceID   int64
	Limit     int
	Offset    int
}

// ActivityMonitor coordinates centralized observability, in-memory ring buffers,
// WebSocket broadcasts, and persistence for the Superadmin System Activity Monitor
type ActivityMonitor struct {
	mu             sync.RWMutex
	events         []ActivityEvent
	maxEvents      int
	nextID         int64
	snapshots      map[uint]*DevicePollingSnapshot
	jobs           map[string]*BackgroundJobSnapshot
	hub            *websocket.Hub
	systemRepo     *repository.SystemRepository
	commTracer     *CommTracer
	startTime      time.Time
	lastActivity   time.Time
	pollingStatus  string // RUNNING, IDLE, STOPPED
	activePollsMap map[uint]time.Time
	pollMapMu      sync.RWMutex
}

var (
	GlobalActivityMonitor *ActivityMonitor
	globalMonitorOnce     sync.Once
)

// NewActivityMonitor constructs the monitor singleton
func NewActivityMonitor(hub *websocket.Hub, systemRepo *repository.SystemRepository, commTracer *CommTracer) *ActivityMonitor {
	m := &ActivityMonitor{
		events:         make([]ActivityEvent, 0, 1000),
		maxEvents:      1000,
		snapshots:      make(map[uint]*DevicePollingSnapshot),
		jobs:           make(map[string]*BackgroundJobSnapshot),
		hub:            hub,
		systemRepo:     systemRepo,
		commTracer:     commTracer,
		startTime:      time.Now(),
		lastActivity:   time.Now(),
		pollingStatus:  "RUNNING",
		activePollsMap: make(map[uint]time.Time),
	}

	// Initialize standard background job slots
	m.initBackgroundJobs()

	globalMonitorOnce.Do(func() {
		GlobalActivityMonitor = m
	})

	return m
}

// GetGlobalActivityMonitor returns the global monitor instance
func GetGlobalActivityMonitor() *ActivityMonitor {
	return GlobalActivityMonitor
}

func (m *ActivityMonitor) initBackgroundJobs() {
	now := time.Now()
	m.jobs["polling_scheduler"] = &BackgroundJobSnapshot{
		JobName:       "polling_scheduler",
		Category:      "COMMUNICATION",
		Status:        "RUNNING",
		LastStarted:   &now,
		LastResult:    "Active per-device polling loops running",
		Details:       "5s standard polling cadence with hardware hot-plug detection",
	}
	m.jobs["aggregation_pipeline"] = &BackgroundJobSnapshot{
		JobName:       "aggregation_pipeline",
		Category:      "DATA_PROCESSING",
		Status:        "RUNNING",
		LastStarted:   &now,
		LastResult:    "1m, 5m, 1h, 24h periodic rollup workers active",
		Details:       "Continuous rollup pipeline triggered upon telemetry ingestion",
	}
	m.jobs["ispu_calculator"] = &BackgroundJobSnapshot{
		JobName:       "ispu_calculator",
		Category:      "ENVIRONMENTAL_INDEX",
		Status:        "RUNNING",
		LastStarted:   &now,
		LastResult:    "Permen LHK No. 14/2020 calculation worker ready",
		Details:       "Enqueued on 24-hour completed aggregation buckets",
	}
	m.jobs["data_retention_housekeeping"] = &BackgroundJobSnapshot{
		JobName:       "data_retention_housekeeping",
		Category:      "STORAGE",
		Status:        "IDLE",
		LastResult:    "Scheduled daily cleanup of stale telemetry > 30 days",
		Details:       "Automated housekeeping keeps database lean",
	}
	m.jobs["backup_scheduler"] = &BackgroundJobSnapshot{
		JobName:       "backup_scheduler",
		Category:      "DISASTER_RECOVERY",
		Status:        "IDLE",
		LastResult:    "Automated backup engine initialized",
		Details:       "Scheduled mysqldump archives",
	}
	m.jobs["websocket_broadcaster"] = &BackgroundJobSnapshot{
		JobName:       "websocket_broadcaster",
		Category:      "TELEMETRY_STREAM",
		Status:        "RUNNING",
		LastStarted:   &now,
		LastResult:    "WebSocket hub active",
		Details:       "Real-time event and telemetry push to connected SCADA clients",
	}
}

// RecordEvent records a structured activity event into the in-memory ring buffer,
// broadcasts it in real-time over WebSocket, and asynchronously logs warnings/errors
// into MariaDB system_logs
func (m *ActivityMonitor) RecordEvent(event ActivityEvent) {
	if m == nil {
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Severity == "" {
		event.Severity = SeverityInfo
	}
	if event.Category == "" {
		event.Category = CategorySystem
	}

	eventID := atomic.AddInt64(&m.nextID, 1)
	event.ID = eventID

	m.mu.Lock()
	m.lastActivity = event.Timestamp

	// Add to in-memory ring buffer
	if len(m.events) >= m.maxEvents {
		// Drop oldest 100 entries to amortize slice re-allocations
		m.events = append(m.events[100:], event)
	} else {
		m.events = append(m.events, event)
	}
	m.mu.Unlock()

	// 1. Asynchronously persist high-severity events to MariaDB system_logs
	if m.systemRepo != nil && (event.Severity == SeverityWarning || event.Severity == SeverityError || event.Severity == SeverityCritical) {
		go func(ev ActivityEvent) {
			details := ev.ErrorDetail
			if details == "" && len(ev.Metadata) > 0 {
				if b, err := json.Marshal(ev.Metadata); err == nil {
					details = string(b)
				}
			}
			_ = m.systemRepo.AddSystemLog(&model.SystemLog{
				Level:     ev.Severity,
				Component: ev.Component,
				Message:   fmt.Sprintf("[%s] %s: %s", ev.Category, ev.EventName, ev.Description),
				Details:   details,
				CreatedAt: ev.Timestamp,
			})
		}(event)
	}

	// 2. Real-time broadcast to connected Superadmin clients over WebSocket
	if m.hub != nil {
		m.hub.Broadcast(websocket.WSMessage{
			Type:      "SYSTEM_ACTIVITY_EVENT",
			Timestamp: event.Timestamp,
			Data:      event,
		})
	}
}

// SetPollingEngineStatus updates the status of the polling engine
func (m *ActivityMonitor) SetPollingEngineStatus(status string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.pollingStatus = status
	m.mu.Unlock()
}

// MarkPollStarted records that a polling cycle has started for a device
func (m *ActivityMonitor) MarkPollStarted(deviceID uint) {
	if m == nil {
		return
	}
	m.pollMapMu.Lock()
	m.activePollsMap[deviceID] = time.Now()
	m.pollMapMu.Unlock()
}

// MarkPollEnded records that a polling cycle has ended for a device
func (m *ActivityMonitor) MarkPollEnded(deviceID uint) {
	if m == nil {
		return
	}
	m.pollMapMu.Lock()
	delete(m.activePollsMap, deviceID)
	m.pollMapMu.Unlock()
}

// UpdateDevicePollingSnapshot updates or initializes a device's polling status snapshot
func (m *ActivityMonitor) UpdateDevicePollingSnapshot(deviceID uint, updateFn func(*DevicePollingSnapshot)) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	snap, exists := m.snapshots[deviceID]
	if !exists {
		snap = &DevicePollingSnapshot{
			DeviceID:            deviceID,
			PollingState:        PollingStateWaiting,
			CommunicationStatus: "UNKNOWN",
			RecoveryState:       "NORMAL",
		}
		m.snapshots[deviceID] = snap
	}

	updateFn(snap)
}

// UpdateBackgroundJob updates the execution state of a background job
func (m *ActivityMonitor) UpdateBackgroundJob(name, category, status string, durationMs int, result, errStr string, nextRun *time.Time, details string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[name]
	if !exists {
		job = &BackgroundJobSnapshot{
			JobName:  name,
			Category: category,
		}
		m.jobs[name] = job
	}

	now := time.Now()
	job.Status = status
	if status == "RUNNING" {
		job.LastStarted = &now
	} else if status == "IDLE" || status == "COMPLETED" || status == "STOPPED" {
		job.LastCompleted = &now
	}
	if durationMs > 0 {
		job.LastDurationMs = durationMs
	}
	if result != "" {
		job.LastResult = result
	}
	if errStr != "" {
		job.LastError = errStr
	}
	if nextRun != nil {
		job.NextRun = nextRun
	}
	if details != "" {
		job.Details = details
	}
}

// GetSystemOverview returns executive summary metrics for the system activity monitor header
func (m *ActivityMonitor) GetSystemOverview(devices []model.Device) SystemOverviewSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	uptime := int64(now.Sub(m.startTime).Seconds())

	onlineCount := 0
	offlineCount := 0
	reconnCount := 0
	errorCount := 0
	timeoutCount := 0

	for _, d := range devices {
		snap, hasSnap := m.snapshots[d.ID]
		commStatus := string(d.ConnectionStatus)
		if hasSnap && snap.CommunicationStatus != "" {
			commStatus = snap.CommunicationStatus
		}

		switch strings.ToUpper(commStatus) {
		case "ONLINE":
			onlineCount++
		case "RECONNECTING", "DEGRADED":
			reconnCount++
		default:
			offlineCount++
		}

		if hasSnap {
			if snap.PollingState == PollingStateTimeout || snap.ConsecutiveFailures > 0 {
				timeoutCount++
			}
			if snap.LastError != "" || snap.FailedParameters > 0 {
				errorCount++
			}
		}
	}

	m.pollMapMu.RLock()
	activePolls := len(m.activePollsMap)
	m.pollMapMu.RUnlock()

	// Overall status evaluation: DEGRADED if any offline/reconnecting or errors, RUNNING if all healthy
	overallStatus := "RUNNING"
	if len(devices) == 0 {
		overallStatus = "IDLE"
	} else if offlineCount > 0 || reconnCount > 0 || errorCount > 0 {
		overallStatus = "DEGRADED"
	}
	if m.pollingStatus == "STOPPED" {
		overallStatus = "STOPPED"
	}

	lastAct := m.lastActivity
	if lastAct.IsZero() {
		lastAct = now
	}

	return SystemOverviewSnapshot{
		OverallStatus:         overallStatus,
		BackendServiceStatus:  "RUNNING",
		PollingEngineStatus:   m.pollingStatus,
		DevicesOnline:         onlineCount,
		DevicesOffline:        offlineCount,
		DevicesReconnecting:   reconnCount,
		TotalDevices:          len(devices),
		ActivePollsCount:      activePolls,
		WaitingOrTimeoutCount: timeoutCount,
		ActiveErrorsCount:     errorCount,
		LastActivityTime:      &lastAct,
		ServerTime:            now,
		UptimeSeconds:         uptime,
	}
}

// GetEvents retrieves filtered activity events from the in-memory ring buffer
func (m *ActivityMonitor) GetEvents(filter ActivityFilter) ([]ActivityEvent, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []ActivityEvent
	searchLower := strings.ToLower(filter.Search)
	sevUpper := strings.ToUpper(filter.Severity)
	catUpper := strings.ToUpper(filter.Category)

	// Iterate backwards (newest first)
	for i := len(m.events) - 1; i >= 0; i-- {
		ev := m.events[i]

		if filter.SinceID > 0 && ev.ID <= filter.SinceID {
			continue
		}
		if filter.DeviceID > 0 && ev.DeviceID != filter.DeviceID {
			continue
		}
		if sevUpper != "" && sevUpper != "ALL" && ev.Severity != sevUpper {
			continue
		}
		if catUpper != "" && catUpper != "ALL" && ev.Category != catUpper {
			continue
		}
		if filter.StartTime != nil && ev.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && ev.Timestamp.After(*filter.EndTime) {
			continue
		}

		if searchLower != "" {
			combined := strings.ToLower(fmt.Sprintf("%s %s %s %s %s %s",
				ev.EventName, ev.Description, ev.Component, ev.DeviceCode, ev.DeviceName, ev.ErrorDetail))
			if !strings.Contains(combined, searchLower) {
				continue
			}
		}

		matched = append(matched, ev)
	}

	total := len(matched)
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return []ActivityEvent{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return matched[offset:end], total
}

// GetAllPollingSnapshots returns current snapshots for all tracked devices
func (m *ActivityMonitor) GetAllPollingSnapshots() []DevicePollingSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]DevicePollingSnapshot, 0, len(m.snapshots))
	for _, snap := range m.snapshots {
		res = append(res, *snap)
	}
	return res
}

// GetDevicePollingSnapshot returns the snapshot for a specific device
func (m *ActivityMonitor) GetDevicePollingSnapshot(deviceID uint) (*DevicePollingSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snap, exists := m.snapshots[deviceID]
	if !exists {
		return nil, false
	}
	cp := *snap
	return &cp, true
}

// GetAllBackgroundJobs returns snapshots of all background jobs
func (m *ActivityMonitor) GetAllBackgroundJobs() []BackgroundJobSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]BackgroundJobSnapshot, 0, len(m.jobs))
	for _, job := range m.jobs {
		res = append(res, *job)
	}
	return res
}

// ClearEvents clears the in-memory ring buffer
func (m *ActivityMonitor) ClearEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = m.events[:0]
	logger.Info("[ActivityMonitor] In-memory activity event ring buffer cleared by Superadmin")
}

// ExportEvents exports filtered activity events to CSV or JSON format
func (m *ActivityMonitor) ExportEvents(filter ActivityFilter, format string) ([]byte, string, error) {
	events, _ := m.GetEvents(filter)

	if strings.ToLower(format) == "json" {
		b, err := json.MarshalIndent(events, "", "  ")
		return b, "application/json", err
	}

	// CSV Export
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write CSV Header
	_ = w.Write([]string{
		"ID", "Timestamp", "Severity", "Category", "Component",
		"Device ID", "Device Code", "Event Name", "Description",
		"Duration (ms)", "Error Detail", "Correlation ID",
	})

	for _, ev := range events {
		_ = w.Write([]string{
			strconv.FormatInt(ev.ID, 10),
			ev.Timestamp.Format(time.RFC3339),
			ev.Severity,
			ev.Category,
			ev.Component,
			strconv.FormatUint(uint64(ev.DeviceID), 10),
			ev.DeviceCode,
			ev.EventName,
			ev.Description,
			strconv.Itoa(ev.DurationMs),
			ev.ErrorDetail,
			ev.CorrelationID,
		})
	}
	w.Flush()

	return buf.Bytes(), "text/csv", nil
}
