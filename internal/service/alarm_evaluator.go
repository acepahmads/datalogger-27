package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
)

const (
	// Wave 1: Threshold Alarm Codes
	AlarmCodeCriticalHigh = "THRESHOLD_CRITICAL_HIGH"
	AlarmCodeWarningHigh  = "THRESHOLD_WARNING_HIGH"
	AlarmCodeWarningLow   = "THRESHOLD_WARNING_LOW"
	AlarmCodeCriticalLow  = "THRESHOLD_CRITICAL_LOW"

	// Wave 2: Device Health, Communication & Data Abnormality Codes
	AlarmCodeDeviceOffline    = "DEVICE_OFFLINE"
	AlarmCodeCommError        = "COMMUNICATION_ERROR"
	AlarmCodeAbnormalityStuck = "DATA_ABNORMALITY_STUCK_VALUE"
	AlarmCodeAbnormalityBad   = "DATA_ABNORMALITY_BAD_QUALITY"
	AlarmCodeAbnormalityStale = "DATA_ABNORMALITY_STALE_DATA"
	AlarmCodeAbnormalitySpike = "DATA_ABNORMALITY_SPIKE"

	DefaultHysteresisPercent        = 1.0 // 1.0% of threshold value by default
	DefaultDeviceOfflineTimeout     = 60 * time.Second
	DefaultStartupGracePeriod       = 30 * time.Second
	DefaultCommFailureThreshold     = 5
	DefaultCommRecoveryThreshold    = 2
	DefaultStuckValueCountThreshold = 10
	DefaultStuckValueDuration       = 30 * time.Second
	DefaultBadQualityCountThreshold = 5
	DefaultStaleDataCountThreshold  = 3
	DefaultStuckTolerance           = 1e-6
)

// ActiveAlarmEntry holds an in-memory cached state of an active/acknowledged alarm
type ActiveAlarmEntry struct {
	AlarmID        uint
	AlarmCode      string
	Severity       model.AlarmSeverity
	Status         model.AlarmStatus
	TriggerValue   float64
	ThresholdValue float64
	TriggeredAt    time.Time
}

// DeviceHealthInput decouples communication health tracking from the evaluator
type DeviceHealthInput struct {
	DeviceID             uint
	State                string
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LastCommSuccess      *time.Time
	LastCommFailure      *time.Time
	LastErrorCategory    string
	LastError            string
}

// paramAbnormalityState tracks rolling behavior per parameter for data abnormality detection
type paramAbnormalityState struct {
	lastValue        float64
	lastValueTime    time.Time
	hasValue         bool
	stuckCount       int
	stuckSince       time.Time
	consecutiveBad   int
	consecutiveStale int
	normalSinceSpike int
}

// AlarmEvaluatorConfig provides runtime configuration for all alarm evaluations
type AlarmEvaluatorConfig struct {
	DefaultHysteresisPercent float64
	AutoReloadActiveAlarms   bool

	// Wave 2 Options
	DeviceOfflineTimeout        time.Duration
	StartupGracePeriod          time.Duration
	CommErrorConsecutiveLimit   int
	CommRecoverySuccessLimit    int
	StuckValueCountThreshold    int
	StuckValueDurationThreshold time.Duration
	BadQualityCountThreshold    int
	StaleDataCountThreshold     int
	StuckTolerance              float64
}

// DefaultAlarmEvaluatorConfig returns standard edge configuration
func DefaultAlarmEvaluatorConfig() AlarmEvaluatorConfig {
	return AlarmEvaluatorConfig{
		DefaultHysteresisPercent:    DefaultHysteresisPercent,
		AutoReloadActiveAlarms:      true,
		DeviceOfflineTimeout:        DefaultDeviceOfflineTimeout,
		StartupGracePeriod:          DefaultStartupGracePeriod,
		CommErrorConsecutiveLimit:   DefaultCommFailureThreshold,
		CommRecoverySuccessLimit:    DefaultCommRecoveryThreshold,
		StuckValueCountThreshold:    DefaultStuckValueCountThreshold,
		StuckValueDurationThreshold: DefaultStuckValueDuration,
		BadQualityCountThreshold:    DefaultBadQualityCountThreshold,
		StaleDataCountThreshold:     DefaultStaleDataCountThreshold,
		StuckTolerance:              DefaultStuckTolerance,
	}
}

// ParameterLookupFunc allows the evaluator to resolve full parameter metadata if omitted in payload
type ParameterLookupFunc func(ctx context.Context, paramID uint) (*model.Parameter, error)

// AlarmEvaluator performs thread-safe threshold, device health, communication error, and data abnormality evaluation
type AlarmEvaluator struct {
	alarmService *AlarmService
	cfg          AlarmEvaluatorConfig
	paramLookup  ParameterLookupFunc
	mu           sync.RWMutex
	startTime    time.Time

	// Wave 1 active alarms: key "deviceID:paramID"
	activeHigh map[string]*ActiveAlarmEntry
	activeLow  map[string]*ActiveAlarmEntry

	// Wave 2 active alarms
	activeDeviceOffline map[uint]*ActiveAlarmEntry        // key: deviceID
	activeCommErrors    map[uint]*ActiveAlarmEntry        // key: deviceID
	activeAbnormality   map[string]*ActiveAlarmEntry      // key: "deviceID:paramID:alarmCode"
	abnormalityTracking map[string]*paramAbnormalityState // key: "deviceID:paramID"

	// Evaluation metrics
	evaluationsCount      uint64
	alarmsTriggeredCount  uint64
	alarmsClearedCount    uint64
	evaluationErrorsCount uint64
}

// NewAlarmEvaluator constructs a new AlarmEvaluator and synchronizes active alarms from DB
func NewAlarmEvaluator(ctx context.Context, alarmService *AlarmService, cfg AlarmEvaluatorConfig, lookup ParameterLookupFunc) *AlarmEvaluator {
	if cfg.DefaultHysteresisPercent <= 0 {
		cfg.DefaultHysteresisPercent = DefaultHysteresisPercent
	}
	if cfg.DeviceOfflineTimeout <= 0 {
		cfg.DeviceOfflineTimeout = DefaultDeviceOfflineTimeout
	}
	if cfg.StartupGracePeriod <= 0 {
		cfg.StartupGracePeriod = DefaultStartupGracePeriod
	}
	if cfg.CommErrorConsecutiveLimit <= 0 {
		cfg.CommErrorConsecutiveLimit = DefaultCommFailureThreshold
	}
	if cfg.CommRecoverySuccessLimit <= 0 {
		cfg.CommRecoverySuccessLimit = DefaultCommRecoveryThreshold
	}
	if cfg.StuckValueCountThreshold <= 0 {
		cfg.StuckValueCountThreshold = DefaultStuckValueCountThreshold
	}
	if cfg.StuckValueDurationThreshold <= 0 {
		cfg.StuckValueDurationThreshold = DefaultStuckValueDuration
	}
	if cfg.BadQualityCountThreshold <= 0 {
		cfg.BadQualityCountThreshold = DefaultBadQualityCountThreshold
	}
	if cfg.StaleDataCountThreshold <= 0 {
		cfg.StaleDataCountThreshold = DefaultStaleDataCountThreshold
	}
	if cfg.StuckTolerance <= 0 {
		cfg.StuckTolerance = DefaultStuckTolerance
	}

	e := &AlarmEvaluator{
		alarmService:        alarmService,
		cfg:                 cfg,
		paramLookup:         lookup,
		startTime:           time.Now().UTC(),
		activeHigh:          make(map[string]*ActiveAlarmEntry),
		activeLow:           make(map[string]*ActiveAlarmEntry),
		activeDeviceOffline: make(map[uint]*ActiveAlarmEntry),
		activeCommErrors:    make(map[uint]*ActiveAlarmEntry),
		activeAbnormality:   make(map[string]*ActiveAlarmEntry),
		abnormalityTracking: make(map[string]*paramAbnormalityState),
	}

	if cfg.AutoReloadActiveAlarms && alarmService != nil {
		if err := e.ReloadActiveAlarms(ctx); err != nil {
			logger.Warn("Failed to reload active alarms on evaluator initialization: %v", err)
		}
	}

	return e
}

// ReloadActiveAlarms queries active and acknowledged alarms from database to populate in-memory state
func (e *AlarmEvaluator) ReloadActiveAlarms(ctx context.Context) error {
	if e.alarmService == nil {
		return nil
	}

	alarms, err := e.alarmService.GetActiveOrAckedAlarms(ctx)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeHigh = make(map[string]*ActiveAlarmEntry)
	e.activeLow = make(map[string]*ActiveAlarmEntry)
	e.activeDeviceOffline = make(map[uint]*ActiveAlarmEntry)
	e.activeCommErrors = make(map[uint]*ActiveAlarmEntry)
	e.activeAbnormality = make(map[string]*ActiveAlarmEntry)

	for i := range alarms {
		alm := alarms[i]
		if alm.DeviceID == nil {
			continue
		}

		entry := &ActiveAlarmEntry{
			AlarmID:     alm.ID,
			AlarmCode:   alm.AlarmCode,
			Severity:    alm.Severity,
			Status:      alm.Status,
			TriggeredAt: alm.TriggeredAt,
		}
		if alm.TriggerValue != nil {
			entry.TriggerValue = *alm.TriggerValue
		}
		if alm.ThresholdValue != nil {
			entry.ThresholdValue = *alm.ThresholdValue
		}

		// Device-level alarms (no parameter_id)
		switch alm.AlarmCode {
		case AlarmCodeDeviceOffline:
			e.activeDeviceOffline[*alm.DeviceID] = entry
			continue
		case AlarmCodeCommError:
			e.activeCommErrors[*alm.DeviceID] = entry
			continue
		}
		if strings.HasPrefix(alm.AlarmCode, "COMMUNICATION_ERROR") {
			e.activeCommErrors[*alm.DeviceID] = entry
			continue
		}

		// Parameter-level alarms
		if alm.ParameterID == nil {
			continue
		}
		key := fmt.Sprintf("%d:%d", *alm.DeviceID, *alm.ParameterID)

		switch alm.AlarmCode {
		case AlarmCodeCriticalHigh, AlarmCodeWarningHigh:
			e.activeHigh[key] = entry
		case AlarmCodeCriticalLow, AlarmCodeWarningLow:
			e.activeLow[key] = entry
		case AlarmCodeAbnormalityStuck, AlarmCodeAbnormalityBad, AlarmCodeAbnormalityStale, AlarmCodeAbnormalitySpike:
			abKey := fmt.Sprintf("%s:%s", key, alm.AlarmCode)
			e.activeAbnormality[abKey] = entry
		}
	}

	totalLoaded := len(e.activeHigh) + len(e.activeLow) + len(e.activeDeviceOffline) + len(e.activeCommErrors) + len(e.activeAbnormality)
	logger.Info("AlarmEvaluator synchronized %d active/acknowledged alarms from database (offline: %d, comm: %d, abnormal: %d, thresh: %d)",
		totalLoaded, len(e.activeDeviceOffline), len(e.activeCommErrors), len(e.activeAbnormality), len(e.activeHigh)+len(e.activeLow))
	return nil
}

// EvaluateTelemetry evaluates processed telemetry against abnormalities and configured thresholds
func (e *AlarmEvaluator) EvaluateTelemetry(ctx context.Context, latest *LatestTelemetry, param *model.Parameter) {
	if latest == nil {
		return
	}

	atomic.AddUint64(&e.evaluationsCount, 1)

	// Resolve Parameter Configuration if needed
	if param == nil && e.paramLookup != nil {
		p, err := e.paramLookup(ctx, latest.ParameterID)
		if err != nil {
			atomic.AddUint64(&e.evaluationErrorsCount, 1)
			logger.Debug("Evaluator parameter lookup failed for ID %d: %v", latest.ParameterID, err)
			return
		}
		param = p
	}

	if param == nil || !param.Enabled {
		return
	}

	key := fmt.Sprintf("%d:%d", latest.DeviceID, latest.ParameterID)

	e.mu.Lock()
	defer e.mu.Unlock()

	// Wave 2: Evaluate Data Abnormality (stuck, bad quality, stale, spike)
	e.evaluateDataAbnormality(ctx, latest, param)

	// Wave 1: Numeric validity checks for Threshold Alarms:
	val := latest.ProcessedValue
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return
	}
	if latest.Quality == model.QualityBad {
		return
	}

	// Wave 1: Threshold evaluation
	e.evaluateHighSide(ctx, key, latest, param, val)
	e.evaluateLowSide(ctx, key, latest, param, val)
}

// evaluateDataAbnormality inspects telemetry for stuck values, bad quality accumulation, stale samples, and spikes
func (e *AlarmEvaluator) evaluateDataAbnormality(ctx context.Context, latest *LatestTelemetry, param *model.Parameter) {
	paramKey := fmt.Sprintf("%d:%d", latest.DeviceID, latest.ParameterID)
	st, exists := e.abnormalityTracking[paramKey]
	if !exists {
		st = &paramAbnormalityState{}
		e.abnormalityTracking[paramKey] = st
	}

	now := latest.ReceivedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	// 1. Repeated BAD Quality Check
	badLimit := e.cfg.BadQualityCountThreshold
	if badLimit <= 0 {
		badLimit = DefaultBadQualityCountThreshold
	}
	badKey := fmt.Sprintf("%s:%s", paramKey, AlarmCodeAbnormalityBad)

	if latest.Quality == model.QualityBad {
		st.consecutiveBad++
		if st.consecutiveBad >= badLimit {
			if _, ok := e.activeAbnormality[badKey]; !ok {
				reason := string(latest.QualityReason)
				if reason == "" {
					reason = "BAD_QUALITY"
				}
				msg := fmt.Sprintf("Parameter %s on device %s has %d consecutive BAD quality samples: %s",
					param.ParameterName, latest.DeviceCode, st.consecutiveBad, reason)
				entry := e.createDeviceParamAlarm(ctx, latest.DeviceID, latest.ParameterID, latest.DeviceCode, param.ParameterCode, AlarmCodeAbnormalityBad, model.SeverityWarning, msg, nil, nil)
				if entry != nil {
					e.activeAbnormality[badKey] = entry
				}
			}
		}
	} else if latest.Quality == model.QualityGood {
		if st.consecutiveBad > 0 {
			st.consecutiveBad = 0
		}
		if existing, ok := e.activeAbnormality[badKey]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeAbnormality, badKey)
		}
	}

	// 2. Repeated Stale Data Check
	staleLimit := e.cfg.StaleDataCountThreshold
	if staleLimit <= 0 {
		staleLimit = DefaultStaleDataCountThreshold
	}
	staleKey := fmt.Sprintf("%s:%s", paramKey, AlarmCodeAbnormalityStale)

	if latest.Quality == model.QualityStale || strings.Contains(latest.QualityFlags, model.FlagStale) {
		st.consecutiveStale++
		if st.consecutiveStale >= staleLimit {
			if _, ok := e.activeAbnormality[staleKey]; !ok {
				msg := fmt.Sprintf("Parameter %s on device %s has repeated STALE data (%d consecutive stale samples)",
					param.ParameterName, latest.DeviceCode, st.consecutiveStale)
				entry := e.createDeviceParamAlarm(ctx, latest.DeviceID, latest.ParameterID, latest.DeviceCode, param.ParameterCode, AlarmCodeAbnormalityStale, model.SeverityWarning, msg, nil, nil)
				if entry != nil {
					e.activeAbnormality[staleKey] = entry
				}
			}
		}
	} else if latest.Quality == model.QualityGood {
		if st.consecutiveStale > 0 {
			st.consecutiveStale = 0
		}
		if existing, ok := e.activeAbnormality[staleKey]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeAbnormality, staleKey)
		}
	}

	// 3. Spike / Outlier Check
	spikeKey := fmt.Sprintf("%s:%s", paramKey, AlarmCodeAbnormalitySpike)
	isSpike := strings.Contains(latest.QualityFlags, model.FlagSpike) || latest.QualityReason == model.ReasonSpikeDetected
	if isSpike {
		st.normalSinceSpike = 0
		if _, ok := e.activeAbnormality[spikeKey]; !ok {
			val := latest.ProcessedValue
			msg := fmt.Sprintf("Parameter %s on device %s detected sudden SPIKE value: %.2f %s",
				param.ParameterName, latest.DeviceCode, val, param.Unit)
			entry := e.createDeviceParamAlarm(ctx, latest.DeviceID, latest.ParameterID, latest.DeviceCode, param.ParameterCode, AlarmCodeAbnormalitySpike, model.SeverityWarning, msg, &val, nil)
			if entry != nil {
				e.activeAbnormality[spikeKey] = entry
			}
		}
	} else {
		if st.normalSinceSpike < 3 {
			st.normalSinceSpike++
		}
		if st.normalSinceSpike >= 3 {
			if existing, ok := e.activeAbnormality[spikeKey]; ok {
				e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
				delete(e.activeAbnormality, spikeKey)
			}
		}
	}

	// 4. Stuck Value Check (Evaluated on engineering telemetry)
	val := latest.ProcessedValue
	if math.IsNaN(val) || math.IsInf(val, 0) || latest.Quality == model.QualityBad {
		return
	}

	stuckKey := fmt.Sprintf("%s:%s", paramKey, AlarmCodeAbnormalityStuck)
	tol := e.cfg.StuckTolerance
	if tol <= 0 {
		tol = DefaultStuckTolerance
	}
	countLimit := e.cfg.StuckValueCountThreshold
	if countLimit <= 0 {
		countLimit = DefaultStuckValueCountThreshold
	}
	durLimit := e.cfg.StuckValueDurationThreshold
	if durLimit <= 0 {
		durLimit = DefaultStuckValueDuration
	}

	if !st.hasValue {
		st.hasValue = true
		st.lastValue = val
		st.lastValueTime = now
		st.stuckCount = 1
		st.stuckSince = now
		return
	}

	diff := math.Abs(val - st.lastValue)
	if diff <= tol {
		st.stuckCount++
		if st.stuckSince.IsZero() {
			st.stuckSince = now
		}
		dur := now.Sub(st.stuckSince)
		// Check both count AND duration criteria before triggering
		if st.stuckCount >= countLimit && dur >= durLimit {
			if _, ok := e.activeAbnormality[stuckKey]; !ok {
				msg := fmt.Sprintf("Parameter %s on device %s has STUCK value: %.2f %s (%d identical readings over %v)",
					param.ParameterName, latest.DeviceCode, val, param.Unit, st.stuckCount, dur.Round(time.Second))
				entry := e.createDeviceParamAlarm(ctx, latest.DeviceID, latest.ParameterID, latest.DeviceCode, param.ParameterCode, AlarmCodeAbnormalityStuck, model.SeverityWarning, msg, &val, nil)
				if entry != nil {
					e.activeAbnormality[stuckKey] = entry
				}
			}
		}
	} else {
		// Value has changed! Reset stuck counter and clear alarm if present
		st.lastValue = val
		st.lastValueTime = now
		st.stuckCount = 1
		st.stuckSince = now
		if existing, ok := e.activeAbnormality[stuckKey]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeAbnormality, stuckKey)
		}
	}
}

// EvaluateDeviceOffline evaluates device communication heartbeat and triggers/resolves DEVICE_OFFLINE alarms
func (e *AlarmEvaluator) EvaluateDeviceOffline(ctx context.Context, device *model.Device, health *DeviceHealthInput) {
	if device == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Administrative Disablement / Maintenance Suppression
	isDisabled := !device.Enabled ||
		device.Status != model.DeviceStatusActive ||
		(device.Connection != nil && !device.Connection.Enabled) ||
		(health != nil && (health.State == "DISABLED" || health.State == "INACTIVE"))

	if isDisabled {
		// Clear existing offline alarm if present
		if existing, ok := e.activeDeviceOffline[device.ID]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeDeviceOffline, device.ID)
		}
		// Also clear comm error if device was disabled
		if commExisting, ok := e.activeCommErrors[device.ID]; ok {
			e.clearActiveAlarm(ctx, commExisting.AlarmID, commExisting.AlarmCode)
			delete(e.activeCommErrors, device.ID)
		}
		return
	}

	// 2. Startup Grace Period Protection
	now := time.Now().UTC()
	hasEverCommunicated := (health != nil && health.LastCommSuccess != nil && !health.LastCommSuccess.IsZero())
	if !hasEverCommunicated {
		if now.Sub(e.startTime) < e.cfg.StartupGracePeriod {
			// Suppress during initial startup/connection grace period
			return
		}
	}

	// 3. Determine Offline State
	timeout := e.cfg.DeviceOfflineTimeout
	if timeout <= 0 {
		timeout = DefaultDeviceOfflineTimeout
	}

	isOffline := false
	var outageDuration time.Duration

	if hasEverCommunicated {
		outageDuration = now.Sub(*health.LastCommSuccess)
		if outageDuration >= timeout {
			isOffline = true
		}
	} else {
		outageDuration = now.Sub(e.startTime)
		if outageDuration >= timeout {
			isOffline = true
		}
	}

	// Explicit state machine check: StateError is offline
	if health != nil && health.State == "ERROR" {
		isOffline = true
	}

	// If health indicates CONNECTED and not timed out, ensure it's not marked offline
	if health != nil && health.State == "CONNECTED" && (!hasEverCommunicated || outageDuration < timeout) {
		isOffline = false
	}

	// 4. State Transitions
	existing := e.activeDeviceOffline[device.ID]

	if isOffline {
		if existing != nil {
			// Outage episode already alarmed; do NOT create duplicate alarm
			return
		}

		lastSuccessStr := "never"
		if hasEverCommunicated {
			lastSuccessStr = health.LastCommSuccess.Format("2006-01-02 15:04:05 UTC")
		}

		cat := "COMMUNICATION_TIMEOUT"
		if health != nil && health.LastErrorCategory != "" {
			cat = health.LastErrorCategory
		}

		msg := fmt.Sprintf("Device %s is OFFLINE: no communication for %v (last success: %s, category: %s)",
			device.DeviceCode, outageDuration.Round(time.Second), lastSuccessStr, cat)

		entry := e.createDeviceAlarm(ctx, device.ID, device.DeviceCode, AlarmCodeDeviceOffline, model.SeverityCritical, msg)
		if entry != nil {
			e.activeDeviceOffline[device.ID] = entry
		}
	} else {
		// Device is healthy/recovered: resolve outage alarm
		if existing != nil {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeDeviceOffline, device.ID)
		}
	}
}

// EvaluateDeviceCommFailure evaluates communication errors against consecutive failure thresholds
func (e *AlarmEvaluator) EvaluateDeviceCommFailure(ctx context.Context, device *model.Device, health *DeviceHealthInput, err error) {
	if device == nil || err == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Suppress if device is administratively disabled
	if !device.Enabled || device.Status != model.DeviceStatusActive || (health != nil && health.State == "DISABLED") {
		return
	}

	consec := 1
	cat := "COMMUNICATION_FAILURE"
	if health != nil {
		if health.ConsecutiveFailures > 0 {
			consec = health.ConsecutiveFailures
		}
		if health.LastErrorCategory != "" {
			cat = health.LastErrorCategory
		}
	}

	limit := e.cfg.CommErrorConsecutiveLimit
	if limit <= 0 {
		limit = DefaultCommFailureThreshold
	}

	// Do not trigger alarm on single or transient failures below threshold
	if consec < limit {
		return
	}

	// Deduplicate: check if alarm is already active for this device
	if existing := e.activeCommErrors[device.ID]; existing != nil {
		return
	}

	msg := fmt.Sprintf("Device %s communication error [%s]: %s (%d consecutive failures)",
		device.DeviceCode, cat, err.Error(), consec)

	entry := e.createDeviceAlarm(ctx, device.ID, device.DeviceCode, AlarmCodeCommError, model.SeverityWarning, msg)
	if entry != nil {
		e.activeCommErrors[device.ID] = entry
	}
}

// EvaluateDeviceCommSuccess handles communication recovery and clears active comm and offline alarms
func (e *AlarmEvaluator) EvaluateDeviceCommSuccess(ctx context.Context, device *model.Device, health *DeviceHealthInput) {
	if device == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Recover Communication Error Alarm
	consecSuccess := 1
	if health != nil && health.ConsecutiveSuccesses > 0 {
		consecSuccess = health.ConsecutiveSuccesses
	}

	recoveryThreshold := e.cfg.CommRecoverySuccessLimit
	if recoveryThreshold <= 0 {
		recoveryThreshold = DefaultCommRecoveryThreshold
	}

	isRecovered := consecSuccess >= recoveryThreshold || (health != nil && health.State == "CONNECTED")
	if isRecovered {
		if existing, ok := e.activeCommErrors[device.ID]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeCommErrors, device.ID)
		}
	}

	// 2. Recover Device Offline Alarm
	if existing, ok := e.activeDeviceOffline[device.ID]; ok {
		e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
		delete(e.activeDeviceOffline, device.ID)
	}
}

// ClearDeviceAlarms clears all active device-level alarms when a device is deleted or disabled
func (e *AlarmEvaluator) ClearDeviceAlarms(ctx context.Context, deviceID uint) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if existing, ok := e.activeDeviceOffline[deviceID]; ok {
		e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
		delete(e.activeDeviceOffline, deviceID)
	}
	if existing, ok := e.activeCommErrors[deviceID]; ok {
		e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
		delete(e.activeCommErrors, deviceID)
	}
}

// evaluateHighSide manages Warning High and Critical High limits with hysteresis
func (e *AlarmEvaluator) evaluateHighSide(ctx context.Context, key string, latest *LatestTelemetry, param *model.Parameter, val float64) {
	critHigh := param.CriticalHigh
	warnHigh := param.WarningHigh

	if critHigh == nil && warnHigh == nil {
		if existing, ok := e.activeHigh[key]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeHigh, key)
		}
		return
	}

	existing := e.activeHigh[key]

	var hCrit, hWarn float64
	if critHigh != nil {
		hCrit = e.computeHysteresis(*critHigh, param.Hysteresis)
	}
	if warnHigh != nil {
		hWarn = e.computeHysteresis(*warnHigh, param.Hysteresis)
	}

	// Case 1: Value satisfies Critical High threshold
	if critHigh != nil && val >= *critHigh {
		if existing == nil {
			entry := e.createAlarm(ctx, latest, param, AlarmCodeCriticalHigh, model.SeverityCritical, val, *critHigh, "exceeded critical high limit")
			if entry != nil {
				e.activeHigh[key] = entry
			}
		} else if existing.AlarmCode == AlarmCodeWarningHigh {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			entry := e.createAlarm(ctx, latest, param, AlarmCodeCriticalHigh, model.SeverityCritical, val, *critHigh, "escalated from warning to critical high limit")
			if entry != nil {
				e.activeHigh[key] = entry
			}
		}
		return
	}

	// Case 2: Value satisfies Warning High threshold (and below Critical High)
	if warnHigh != nil && val >= *warnHigh {
		if existing == nil {
			entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningHigh, model.SeverityWarning, val, *warnHigh, "exceeded warning high limit")
			if entry != nil {
				e.activeHigh[key] = entry
			}
		} else if existing.AlarmCode == AlarmCodeCriticalHigh {
			if critHigh != nil && val < (*critHigh-hCrit) {
				e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
				entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningHigh, model.SeverityWarning, val, *warnHigh, "de-escalated from critical to warning high limit")
				if entry != nil {
					e.activeHigh[key] = entry
				}
			}
		}
		return
	}

	// Case 3: Value is below Warning High threshold (Normal or Hysteresis recovery)
	if existing != nil {
		if existing.AlarmCode == AlarmCodeCriticalHigh {
			if critHigh != nil && val < (*critHigh-hCrit) {
				if warnHigh == nil || val < (*warnHigh-hWarn) {
					e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
					delete(e.activeHigh, key)
				} else {
					e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
					entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningHigh, model.SeverityWarning, val, *warnHigh, "de-escalated from critical to warning high limit")
					if entry != nil {
						e.activeHigh[key] = entry
					}
				}
			}
		} else if existing.AlarmCode == AlarmCodeWarningHigh {
			if warnHigh != nil && val < (*warnHigh-hWarn) {
				e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
				delete(e.activeHigh, key)
			}
		}
	}
}

// evaluateLowSide manages Warning Low and Critical Low limits with hysteresis
func (e *AlarmEvaluator) evaluateLowSide(ctx context.Context, key string, latest *LatestTelemetry, param *model.Parameter, val float64) {
	critLow := param.CriticalLow
	warnLow := param.WarningLow

	if critLow == nil && warnLow == nil {
		if existing, ok := e.activeLow[key]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeLow, key)
		}
		return
	}

	existing := e.activeLow[key]

	var hCrit, hWarn float64
	if critLow != nil {
		hCrit = e.computeHysteresis(*critLow, param.Hysteresis)
	}
	if warnLow != nil {
		hWarn = e.computeHysteresis(*warnLow, param.Hysteresis)
	}

	// Case 1: Value satisfies Critical Low threshold
	if critLow != nil && val <= *critLow {
		if existing == nil {
			entry := e.createAlarm(ctx, latest, param, AlarmCodeCriticalLow, model.SeverityCritical, val, *critLow, "dropped below critical low limit")
			if entry != nil {
				e.activeLow[key] = entry
			}
		} else if existing.AlarmCode == AlarmCodeWarningLow {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			entry := e.createAlarm(ctx, latest, param, AlarmCodeCriticalLow, model.SeverityCritical, val, *critLow, "escalated from warning to critical low limit")
			if entry != nil {
				e.activeLow[key] = entry
			}
		}
		return
	}

	// Case 2: Value satisfies Warning Low threshold (and above Critical Low)
	if warnLow != nil && val <= *warnLow {
		if existing == nil {
			entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningLow, model.SeverityWarning, val, *warnLow, "dropped below warning low limit")
			if entry != nil {
				e.activeLow[key] = entry
			}
		} else if existing.AlarmCode == AlarmCodeCriticalLow {
			if critLow != nil && val > (*critLow+hCrit) {
				e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
				entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningLow, model.SeverityWarning, val, *warnLow, "de-escalated from critical to warning low limit")
				if entry != nil {
					e.activeLow[key] = entry
				}
			}
		}
		return
	}

	// Case 3: Value is above Warning Low threshold (Normal or Hysteresis recovery)
	if existing != nil {
		if existing.AlarmCode == AlarmCodeCriticalLow {
			if critLow != nil && val > (*critLow+hCrit) {
				if warnLow == nil || val > (*warnLow+hWarn) {
					e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
					delete(e.activeLow, key)
				} else {
					e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
					entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningLow, model.SeverityWarning, val, *warnLow, "de-escalated from critical to warning low limit")
					if entry != nil {
						e.activeLow[key] = entry
					}
				}
			}
		} else if existing.AlarmCode == AlarmCodeWarningLow {
			if warnLow != nil && val > (*warnLow+hWarn) {
				e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
				delete(e.activeLow, key)
			}
		}
	}
}

// computeHysteresis calculates absolute deadband distance from threshold
func (e *AlarmEvaluator) computeHysteresis(threshold float64, customHysteresis *float64) float64 {
	if customHysteresis != nil && *customHysteresis > 0 {
		return *customHysteresis
	}
	pct := e.cfg.DefaultHysteresisPercent
	if pct <= 0 {
		pct = DefaultHysteresisPercent
	}
	h := math.Abs(threshold) * (pct / 100.0)
	if h <= 0.0001 {
		h = 0.5 // Default minimum deadband if threshold is 0
	}
	return h
}

// createAlarm constructs, persists, and audits a threshold alarm record
func (e *AlarmEvaluator) createAlarm(
	ctx context.Context,
	latest *LatestTelemetry,
	param *model.Parameter,
	code string,
	severity model.AlarmSeverity,
	trigVal float64,
	threshVal float64,
	actionDesc string,
) *ActiveAlarmEntry {
	if e.alarmService == nil {
		return nil
	}

	unitStr := param.Unit
	if unitStr == "" {
		unitStr = latest.Unit
	}

	msg := fmt.Sprintf("%s (%s:%s) %s: %.2f %s (Threshold: %.2f %s)",
		param.ParameterName, latest.DeviceCode, param.ParameterCode, actionDesc, trigVal, unitStr, threshVal, unitStr)

	if latest.Quality == model.QualityUncertain {
		msg += fmt.Sprintf(" [Quality: UNCERTAIN - %s]", latest.QualityReason)
	}

	now := time.Now().UTC()
	alarm := &model.Alarm{
		DeviceID:       &latest.DeviceID,
		ParameterID:    &latest.ParameterID,
		AlarmCode:      code,
		Severity:       severity,
		Message:        msg,
		TriggerValue:   &trigVal,
		ThresholdValue: &threshVal,
		Status:         model.AlarmActive,
		TriggeredAt:    now,
	}

	if err := e.alarmService.CreateAlarm(ctx, alarm); err != nil {
		atomic.AddUint64(&e.evaluationErrorsCount, 1)
		logger.Error("AlarmEvaluator failed persisting alarm %s for dev %d param %d: %v", code, latest.DeviceID, latest.ParameterID, err)
		return nil
	}

	atomic.AddUint64(&e.alarmsTriggeredCount, 1)
	logger.Warn("ALARM TRIGGERED: [%s] %s", severity, msg)

	return &ActiveAlarmEntry{
		AlarmID:        alarm.ID,
		AlarmCode:      code,
		Severity:       severity,
		Status:         model.AlarmActive,
		TriggerValue:   trigVal,
		ThresholdValue: threshVal,
		TriggeredAt:    now,
	}
}

// createDeviceAlarm persists a device-level alarm (such as DEVICE_OFFLINE or COMMUNICATION_ERROR)
func (e *AlarmEvaluator) createDeviceAlarm(
	ctx context.Context,
	deviceID uint,
	deviceCode string,
	code string,
	severity model.AlarmSeverity,
	message string,
) *ActiveAlarmEntry {
	if e.alarmService == nil {
		return nil
	}

	now := time.Now().UTC()
	alarm := &model.Alarm{
		DeviceID:    &deviceID,
		AlarmCode:   code,
		Severity:    severity,
		Message:     message,
		Status:      model.AlarmActive,
		TriggeredAt: now,
	}

	if err := e.alarmService.CreateAlarm(ctx, alarm); err != nil {
		atomic.AddUint64(&e.evaluationErrorsCount, 1)
		logger.Error("AlarmEvaluator failed creating device alarm %s for device %d: %v", code, deviceID, err)
		return nil
	}

	atomic.AddUint64(&e.alarmsTriggeredCount, 1)
	logger.Warn("ALARM TRIGGERED: [%s] %s", severity, message)

	return &ActiveAlarmEntry{
		AlarmID:     alarm.ID,
		AlarmCode:   code,
		Severity:    severity,
		Status:      model.AlarmActive,
		TriggeredAt: now,
	}
}

// createDeviceParamAlarm persists a parameter-specific abnormality alarm
func (e *AlarmEvaluator) createDeviceParamAlarm(
	ctx context.Context,
	deviceID, paramID uint,
	deviceCode, paramCode string,
	code string,
	severity model.AlarmSeverity,
	message string,
	trigVal, threshVal *float64,
) *ActiveAlarmEntry {
	if e.alarmService == nil {
		return nil
	}

	now := time.Now().UTC()
	alarm := &model.Alarm{
		DeviceID:       &deviceID,
		ParameterID:    &paramID,
		AlarmCode:      code,
		Severity:       severity,
		Message:        message,
		TriggerValue:   trigVal,
		ThresholdValue: threshVal,
		Status:         model.AlarmActive,
		TriggeredAt:    now,
	}

	if err := e.alarmService.CreateAlarm(ctx, alarm); err != nil {
		atomic.AddUint64(&e.evaluationErrorsCount, 1)
		logger.Error("AlarmEvaluator failed creating parameter alarm %s for dev %d param %d: %v", code, deviceID, paramID, err)
		return nil
	}

	atomic.AddUint64(&e.alarmsTriggeredCount, 1)
	logger.Warn("ALARM TRIGGERED: [%s] %s", severity, message)

	entry := &ActiveAlarmEntry{
		AlarmID:     alarm.ID,
		AlarmCode:   code,
		Severity:    severity,
		Status:      model.AlarmActive,
		TriggeredAt: now,
	}
	if trigVal != nil {
		entry.TriggerValue = *trigVal
	}
	if threshVal != nil {
		entry.ThresholdValue = *threshVal
	}
	return entry
}

// clearActiveAlarm transitions an active alarm to CLEARED state
func (e *AlarmEvaluator) clearActiveAlarm(ctx context.Context, alarmID uint, code string) {
	if e.alarmService == nil || alarmID == 0 {
		return
	}

	_, err := e.alarmService.ClearAlarm(ctx, alarmID)
	if err != nil {
		atomic.AddUint64(&e.evaluationErrorsCount, 1)
		logger.Warn("AlarmEvaluator failed clearing alarm %d (%s): %v", alarmID, code, err)
		return
	}

	atomic.AddUint64(&e.alarmsClearedCount, 1)
	logger.Info("ALARM RECOVERED / CLEARED: ID=%d Code=%s", alarmID, code)
}

// GetMetrics returns real-time evaluation statistics
func (e *AlarmEvaluator) GetMetrics() (evaluations, triggered, cleared, errors uint64) {
	return atomic.LoadUint64(&e.evaluationsCount),
		atomic.LoadUint64(&e.alarmsTriggeredCount),
		atomic.LoadUint64(&e.alarmsClearedCount),
		atomic.LoadUint64(&e.evaluationErrorsCount)
}

// GetActiveCount returns count of all active alarms tracked in memory across all categories
func (e *AlarmEvaluator) GetActiveCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.activeHigh) + len(e.activeLow) + len(e.activeDeviceOffline) + len(e.activeCommErrors) + len(e.activeAbnormality)
}

// GetActiveDeviceOfflineCount returns count of active device offline alarms
func (e *AlarmEvaluator) GetActiveDeviceOfflineCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.activeDeviceOffline)
}

// GetActiveCommErrorsCount returns count of active communication error alarms
func (e *AlarmEvaluator) GetActiveCommErrorsCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.activeCommErrors)
}

// GetActiveAbnormalityCount returns count of active data abnormality alarms
func (e *AlarmEvaluator) GetActiveAbnormalityCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.activeAbnormality)
}
