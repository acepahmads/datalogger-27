package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
)

const (
	AlarmCodeCriticalHigh = "THRESHOLD_CRITICAL_HIGH"
	AlarmCodeWarningHigh  = "THRESHOLD_WARNING_HIGH"
	AlarmCodeWarningLow   = "THRESHOLD_WARNING_LOW"
	AlarmCodeCriticalLow  = "THRESHOLD_CRITICAL_LOW"

	DefaultHysteresisPercent = 1.0 // 1.0% of threshold value by default
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

// AlarmEvaluatorConfig provides runtime configuration for threshold evaluations
type AlarmEvaluatorConfig struct {
	DefaultHysteresisPercent float64
	AutoReloadActiveAlarms   bool
}

// DefaultAlarmEvaluatorConfig returns standard edge configuration
func DefaultAlarmEvaluatorConfig() AlarmEvaluatorConfig {
	return AlarmEvaluatorConfig{
		DefaultHysteresisPercent: DefaultHysteresisPercent,
		AutoReloadActiveAlarms:   true,
	}
}

// ParameterLookupFunc allows the evaluator to resolve full parameter metadata if omitted in payload
type ParameterLookupFunc func(ctx context.Context, paramID uint) (*model.Parameter, error)

// AlarmEvaluator performs thread-safe, hysteresis-aware threshold evaluation on telemetry
type AlarmEvaluator struct {
	alarmService    *AlarmService
	cfg             AlarmEvaluatorConfig
	paramLookup     ParameterLookupFunc
	mu              sync.RWMutex
	activeHigh      map[string]*ActiveAlarmEntry // Key: "deviceID:paramID"
	activeLow       map[string]*ActiveAlarmEntry // Key: "deviceID:paramID"

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

	e := &AlarmEvaluator{
		alarmService: alarmService,
		cfg:          cfg,
		paramLookup:  lookup,
		activeHigh:   make(map[string]*ActiveAlarmEntry),
		activeLow:    make(map[string]*ActiveAlarmEntry),
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

	for i := range alarms {
		alm := alarms[i]
		if alm.DeviceID == nil || alm.ParameterID == nil {
			continue
		}

		key := fmt.Sprintf("%d:%d", *alm.DeviceID, *alm.ParameterID)
		entry := &ActiveAlarmEntry{
			AlarmID:        alm.ID,
			AlarmCode:      alm.AlarmCode,
			Severity:       alm.Severity,
			Status:         alm.Status,
			TriggeredAt:    alm.TriggeredAt,
		}
		if alm.TriggerValue != nil {
			entry.TriggerValue = *alm.TriggerValue
		}
		if alm.ThresholdValue != nil {
			entry.ThresholdValue = *alm.ThresholdValue
		}

		switch alm.AlarmCode {
		case AlarmCodeCriticalHigh, AlarmCodeWarningHigh:
			e.activeHigh[key] = entry
		case AlarmCodeCriticalLow, AlarmCodeWarningLow:
			e.activeLow[key] = entry
		}
	}

	logger.Info("AlarmEvaluator initialized with %d active/acknowledged threshold alarms from database", len(e.activeHigh)+len(e.activeLow))
	return nil
}

// EvaluateTelemetry evaluates processed telemetry against configured thresholds
func (e *AlarmEvaluator) EvaluateTelemetry(ctx context.Context, latest *LatestTelemetry, param *model.Parameter) {
	if latest == nil {
		return
	}

	atomic.AddUint64(&e.evaluationsCount, 1)

	// 1. Data Quality Checks:
	// Do not evaluate numeric threshold alarms from NaN or Infinite values
	val := latest.ProcessedValue
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return
	}

	// Do not trigger new threshold alarms on BAD quality (sensor/decoding failure)
	if latest.Quality == model.QualityBad {
		return
	}

	// 2. Resolve Parameter Configuration:
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

	// Concurrency guard: serialize evaluations for this device:parameter
	e.mu.Lock()
	defer e.mu.Unlock()

	e.evaluateHighSide(ctx, key, latest, param, val)
	e.evaluateLowSide(ctx, key, latest, param, val)
}

// evaluateHighSide manages Warning High and Critical High limits with hysteresis
func (e *AlarmEvaluator) evaluateHighSide(ctx context.Context, key string, latest *LatestTelemetry, param *model.Parameter, val float64) {
	critHigh := param.CriticalHigh
	warnHigh := param.WarningHigh

	// No high limits configured
	if critHigh == nil && warnHigh == nil {
		// If an active high alarm exists from prior config, clear it
		if existing, ok := e.activeHigh[key]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeHigh, key)
		}
		return
	}

	existing := e.activeHigh[key]

	// Calculate hysteresis deadbands
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
			// Trigger new Critical High alarm
			entry := e.createAlarm(ctx, latest, param, AlarmCodeCriticalHigh, model.SeverityCritical, val, *critHigh, "exceeded critical high limit")
			if entry != nil {
				e.activeHigh[key] = entry
			}
		} else if existing.AlarmCode == AlarmCodeWarningHigh {
			// Escalate from Warning High to Critical High
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			entry := e.createAlarm(ctx, latest, param, AlarmCodeCriticalHigh, model.SeverityCritical, val, *critHigh, "escalated from warning to critical high limit")
			if entry != nil {
				e.activeHigh[key] = entry
			}
		}
		// If existing is already Critical High, maintain latched state (no duplicate)
		return
	}

	// Case 2: Value satisfies Warning High threshold (and below Critical High)
	if warnHigh != nil && val >= *warnHigh {
		if existing == nil {
			// Trigger new Warning High alarm
			entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningHigh, model.SeverityWarning, val, *warnHigh, "exceeded warning high limit")
			if entry != nil {
				e.activeHigh[key] = entry
			}
		} else if existing.AlarmCode == AlarmCodeCriticalHigh {
			// Check if Critical High condition recovered past hysteresis
			if critHigh != nil && val < (*critHigh - hCrit) {
				// De-escalate from Critical to Warning High
				e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
				entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningHigh, model.SeverityWarning, val, *warnHigh, "de-escalated from critical to warning high limit")
				if entry != nil {
					e.activeHigh[key] = entry
				}
			}
			// Otherwise still in critical hysteresis band; maintain critical state
		}
		// If existing is already Warning High, maintain latched state (no duplicate)
		return
	}

	// Case 3: Value is below Warning High threshold (Normal or Hysteresis recovery)
	if existing != nil {
		if existing.AlarmCode == AlarmCodeCriticalHigh {
			// Check recovery from Critical High
			if critHigh != nil && val < (*critHigh - hCrit) {
				// Critical recovered. Check if value is also below warning hysteresis
				if warnHigh == nil || val < (*warnHigh - hWarn) {
					// Fully recovered
					e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
					delete(e.activeHigh, key)
				} else {
					// In Warning territory past hysteresis
					e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
					entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningHigh, model.SeverityWarning, val, *warnHigh, "de-escalated from critical to warning high limit")
					if entry != nil {
						e.activeHigh[key] = entry
					}
				}
			}
		} else if existing.AlarmCode == AlarmCodeWarningHigh {
			// Check recovery from Warning High past hysteresis deadband
			if warnHigh != nil && val < (*warnHigh - hWarn) {
				// Fully recovered
				e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
				delete(e.activeHigh, key)
			}
			// If between (warnHigh - hWarn) and warnHigh, stay in latched Warning state
		}
	}
}

// evaluateLowSide manages Warning Low and Critical Low limits with hysteresis
func (e *AlarmEvaluator) evaluateLowSide(ctx context.Context, key string, latest *LatestTelemetry, param *model.Parameter, val float64) {
	critLow := param.CriticalLow
	warnLow := param.WarningLow

	// No low limits configured
	if critLow == nil && warnLow == nil {
		if existing, ok := e.activeLow[key]; ok {
			e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
			delete(e.activeLow, key)
		}
		return
	}

	existing := e.activeLow[key]

	// Calculate hysteresis deadbands
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
			// Escalate from Warning Low to Critical Low
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
			// Check if Critical Low recovered past hysteresis
			if critLow != nil && val > (*critLow + hCrit) {
				// De-escalate to Warning Low
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
			// Check recovery from Critical Low
			if critLow != nil && val > (*critLow + hCrit) {
				if warnLow == nil || val > (*warnLow + hWarn) {
					// Fully recovered
					e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
					delete(e.activeLow, key)
				} else {
					// In Warning Low territory past critical hysteresis
					e.clearActiveAlarm(ctx, existing.AlarmID, existing.AlarmCode)
					entry := e.createAlarm(ctx, latest, param, AlarmCodeWarningLow, model.SeverityWarning, val, *warnLow, "de-escalated from critical to warning low limit")
					if entry != nil {
						e.activeLow[key] = entry
					}
				}
			}
		} else if existing.AlarmCode == AlarmCodeWarningLow {
			// Check recovery from Warning Low past hysteresis deadband
			if warnLow != nil && val > (*warnLow + hWarn) {
				// Fully recovered
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

// createAlarm constructs, persists, and audits an alarm record
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

	// Document uncertain quality if applicable
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

// GetActiveCount returns count of active threshold alarms tracked in memory
func (e *AlarmEvaluator) GetActiveCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.activeHigh) + len(e.activeLow)
}
