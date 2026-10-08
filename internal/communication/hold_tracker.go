package communication

import (
	"fmt"
	"math"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/pkg/formula"
)

// ParameterHoldState tracks the historical state and anomaly grace period for a parameter
type ParameterHoldState struct {
	LastGoodValue     *float64
	LastGoodRaw       *float64
	LastGoodTimestamp time.Time
	AnomalyStartTime  *time.Time
	IsHeld            bool
}

// AnomalyHoldTracker manages hold-last-good-value debounce windows and formula evaluation
type AnomalyHoldTracker struct {
	mu     sync.RWMutex
	states map[uint]*ParameterHoldState
}

// NewAnomalyHoldTracker initializes a new hold tracker
func NewAnomalyHoldTracker() *AnomalyHoldTracker {
	return &AnomalyHoldTracker{
		states: make(map[uint]*ParameterHoldState),
	}
}

// ProcessReading evaluates sensor readings against anomalies, hold-last-value grace period, and custom formulas
func (t *AnomalyHoldTracker) ProcessReading(
	param *model.Parameter,
	rawVal float64,
	engVal float64,
	readErr error,
) (effectiveVal float64, formulaVal *float64, isHeld bool, quality model.TelemetryQuality, err error) {
	if param == nil {
		return engVal, nil, false, model.QualityBad, fmt.Errorf("parameter cannot be nil")
	}

	t.mu.Lock()
	state, exists := t.states[param.ID]
	if !exists {
		state = &ParameterHoldState{}
		t.states[param.ID] = state
	}
	t.mu.Unlock()

	// 1. Determine if reading is anomalous
	isAnomaly := (readErr != nil) ||
		math.IsNaN(engVal) ||
		math.IsInf(engVal, 0) ||
		(param.MinValue != nil && engVal < *param.MinValue) ||
		(param.MaxValue != nil && engVal > *param.MaxValue)

	// 2. If hold-last-value is disabled, process normally
	if !param.HoldLastValueEnabled {
		if readErr != nil {
			return 0, nil, false, model.QualityBad, readErr
		}

		quality = model.QualityGood
		if isAnomaly {
			quality = model.QualityUncertain
		}

		formulaVal = evalParamFormula(param.Formula, engVal, rawVal)
		return engVal, formulaVal, false, quality, nil
	}

	// 3. Hold-last-value is ENABLED:
	holdDuration := time.Duration(param.HoldLastValueSeconds) * time.Second
	if holdDuration <= 0 {
		holdDuration = 120 * time.Second // Default 2 minutes
	}

	now := time.Now()

	// Case A: Reading is NORMAL (good response, within expected bounds)
	if !isAnomaly {
		t.mu.Lock()
		if state.IsHeld {
			logger.Info("Parameter %s (ID %d) sensor recovered to normal: %.2f (clearing hold grace period)",
				param.ParameterCode, param.ID, engVal)
		}
		state.AnomalyStartTime = nil
		state.IsHeld = false
		valCopy := engVal
		rawCopy := rawVal
		state.LastGoodValue = &valCopy
		state.LastGoodRaw = &rawCopy
		state.LastGoodTimestamp = now
		t.mu.Unlock()

		formulaVal = evalParamFormula(param.Formula, engVal, rawVal)
		return engVal, formulaVal, false, model.QualityGood, nil
	}

	// Case B: Reading is ANOMALOUS (timeout, read error, or value outside min/max bounds)
	t.mu.Lock()
	defer t.mu.Unlock()

	// If no previous good value exists yet, must report real state
	if state.LastGoodValue == nil {
		if readErr != nil {
			return 0, nil, false, model.QualityBad, readErr
		}
		formulaVal = evalParamFormula(param.Formula, engVal, rawVal)
		return engVal, formulaVal, false, model.QualityBad, nil
	}

	// First time encountering anomaly? Start grace timer
	if state.AnomalyStartTime == nil {
		state.AnomalyStartTime = &now
		logger.Warn("Parameter %s (ID %d) anomaly detected (err: %v, val: %.2f). Initiating %v hold-last-value grace period.",
			param.ParameterCode, param.ID, readErr, engVal, holdDuration)
	}

	elapsed := now.Sub(*state.AnomalyStartTime)

	// Within grace period (e.g. < 2 minutes): HOLD LAST GOOD VALUE
	if elapsed < holdDuration {
		effectiveVal = *state.LastGoodValue
		effectiveRaw := 0.0
		if state.LastGoodRaw != nil {
			effectiveRaw = *state.LastGoodRaw
		}
		state.IsHeld = true
		formulaVal = evalParamFormula(param.Formula, effectiveVal, effectiveRaw)

		// Return last good value with UNCERTAIN quality and isHeld=true
		return effectiveVal, formulaVal, true, model.QualityUncertain, nil
	}

	// Grace period EXPIRED (>= 2 minutes persistent anomaly): Commit real sensor value
	state.IsHeld = false
	logger.Warn("Parameter %s (ID %d) anomaly persisted for %v (exceeded %v limit). Committing real sensor value to database.",
		param.ParameterCode, param.ID, elapsed.Round(time.Second), holdDuration)

	if readErr != nil {
		return 0, nil, false, model.QualityBad, readErr
	}

	formulaVal = evalParamFormula(param.Formula, engVal, rawVal)
	return engVal, formulaVal, false, model.QualityBad, nil
}

// Reset clears hold state for a specific parameter
func (t *AnomalyHoldTracker) Reset(paramID uint) {
	t.mu.Lock()
	delete(t.states, paramID)
	t.mu.Unlock()
}

// GetState retrieves copy of hold state for a parameter
func (t *AnomalyHoldTracker) GetState(paramID uint) *ParameterHoldState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if s, ok := t.states[paramID]; ok {
		cp := *s
		return &cp
	}
	return nil
}

func evalParamFormula(form string, engVal, rawVal float64) *float64 {
	if form == "" {
		return nil
	}
	vars := map[string]float64{
		"x":     engVal,
		"val":   engVal,
		"value": engVal,
		"raw":   rawVal,
	}
	res, err := formula.Evaluate(form, vars)
	if err != nil {
		logger.Warn("Formula evaluation failed for '%s' (x=%.2f, raw=%.2f): %v", form, engVal, rawVal, err)
		return nil
	}
	return &res
}
