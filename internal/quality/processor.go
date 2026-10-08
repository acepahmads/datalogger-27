package quality

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"datalogger/internal/model"
)

// ValidationInput encapsulates telemetry attributes required for Phase 3.2 quality evaluation
type ValidationInput struct {
	DeviceID        uint
	ParameterID     uint
	RawValue        float64
	Value           float64 // Engineering value (scaled + formula if present)
	DataType        model.ParameterDataType
	DeviceTimestamp *time.Time
	ReceivedAt      time.Time
	Source          string
	IsHeldValue     bool
	ReadError       error
}

// ValidationResult encapsulates the deterministic quality evaluation output
type ValidationResult struct {
	Quality        model.TelemetryQuality
	QualityReason  model.QualityReason
	QualityFlags   string // Comma-separated flags
	ProcessedValue float64
	ProcessedAt    time.Time
	IsStale        bool
	IsSpike        bool
	IsDuplicate    bool
}

// historyPoint records timestamp and value for lightweight spike and duplicate detection
type historyPoint struct {
	Timestamp time.Time
	Value     float64
	RawValue  float64
}

// paramHistory holds a bounded rolling history for a single parameter
type paramHistory struct {
	mu            sync.RWMutex
	points        []historyPoint
	lastValidTime time.Time
	lastRecTime   time.Time
	lastValue     float64
}

// QualityProcessor manages deterministic, isolated telemetry validation and normalization
type QualityProcessor struct {
	historyMu sync.RWMutex
	history   map[uint]*paramHistory // Key: parameter_id
}

// NewQualityProcessor constructs a fresh thread-safe quality processor
func NewQualityProcessor() *QualityProcessor {
	return &QualityProcessor{
		history: make(map[uint]*paramHistory),
	}
}

// getOrCreateHistory retrieves or instantiates the bounded parameter history
func (qp *QualityProcessor) getOrCreateHistory(paramID uint) *paramHistory {
	qp.historyMu.RLock()
	h, exists := qp.history[paramID]
	qp.historyMu.RUnlock()
	if exists {
		return h
	}

	qp.historyMu.Lock()
	defer qp.historyMu.Unlock()
	if h, exists := qp.history[paramID]; exists {
		return h
	}
	h = &paramHistory{
		points: make([]historyPoint, 0, 10),
	}
	qp.history[paramID] = h
	return h
}

// Validate evaluates data quality against configured limits, spikes, timestamps, and anomalies
func (qp *QualityProcessor) Validate(param *model.Parameter, in *ValidationInput) *ValidationResult {
	now := time.Now().UTC()
	recAt := in.ReceivedAt
	if recAt.IsZero() {
		recAt = now
	} else {
		recAt = recAt.UTC()
	}

	res := &ValidationResult{
		Quality:        model.QualityGood,
		QualityReason:  model.ReasonNone,
		ProcessedValue: in.Value,
		ProcessedAt:    now,
	}

	var flags []string
	addFlag := func(f string) {
		for _, existing := range flags {
			if existing == f {
				return
			}
		}
		flags = append(flags, f)
	}

	// 1. Explicit Read / Decode / Communication Error
	if in.ReadError != nil {
		if in.IsHeldValue {
			res.Quality = model.QualityUncertain
			res.QualityReason = model.ReasonHoldAnomalyGrace
			addFlag(model.FlagHeldAnomaly)
		} else {
			res.Quality = model.QualityBad
			res.QualityReason = model.ReasonCommunicationError
			addFlag(model.FlagCommError)
		}
		res.QualityFlags = strings.Join(flags, ",")
		return res
	}

	// 2. NULL / Invalid Numeric / NaN / Infinity Handling
	if math.IsNaN(in.Value) {
		res.Quality = model.QualityBad
		res.QualityReason = model.ReasonNaNValue
		res.ProcessedValue = 0
		addFlag(model.FlagRangeViolation)
		res.QualityFlags = strings.Join(flags, ",")
		return res
	}
	if math.IsInf(in.Value, 0) {
		res.Quality = model.QualityBad
		res.QualityReason = model.ReasonInfiniteValue
		res.ProcessedValue = 0
		addFlag(model.FlagRangeViolation)
		res.QualityFlags = strings.Join(flags, ",")
		return res
	}

	// 3. Timestamp Validation (Future & Excessively Old)
	// Tolerance: 60 seconds into future allowed for minor clock drift, >60s is UNCERTAIN
	if recAt.After(now.Add(60 * time.Second)) {
		res.Quality = model.QualityUncertain
		res.QualityReason = model.ReasonFutureTimestamp
		addFlag(model.FlagTimestampWarn)
	} else if recAt.Before(now.Add(-7 * 24 * time.Hour)) {
		// Older than 7 days is considered excessively old
		res.Quality = model.QualityUncertain
		res.QualityReason = model.ReasonExcessiveOldTimestamp
		addFlag(model.FlagTimestampWarn)
	}

	// 4. Duplicate Data Detection
	hist := qp.getOrCreateHistory(in.ParameterID)
	hist.mu.Lock()
	defer hist.mu.Unlock()

	if len(hist.points) > 0 {
		lastPt := hist.points[len(hist.points)-1]
		if lastPt.Timestamp.Equal(recAt) && lastPt.Value == in.Value {
			res.IsDuplicate = true
			addFlag(model.FlagDuplicate)
			if res.QualityReason == model.ReasonNone {
				res.QualityReason = model.ReasonDuplicateData
			}
		}
	}

	// 5. Range Validation (Hard & Warning limits)
	validationEnabled := true
	if param != nil {
		validationEnabled = param.QualityValidationEnabled
	}

	if validationEnabled && param != nil {
		// Hard Minimum
		minVal := param.MinValue
		if minVal == nil && param.LowLimit != nil {
			minVal = param.LowLimit
		}
		if minVal != nil && in.Value < *minVal {
			res.Quality = model.QualityBad
			res.QualityReason = model.ReasonOutOfHardRange
			addFlag(model.FlagRangeViolation)
		}

		// Hard Maximum
		maxVal := param.MaxValue
		if maxVal == nil && param.HighLimit != nil {
			maxVal = param.HighLimit
		}
		if maxVal != nil && in.Value > *maxVal {
			res.Quality = model.QualityBad
			res.QualityReason = model.ReasonOutOfHardRange
			addFlag(model.FlagRangeViolation)
		}

		// Warning Minimum (Soft Limit)
		if res.Quality != model.QualityBad && param.WarningLow != nil && in.Value < *param.WarningLow {
			res.Quality = model.QualityUncertain
			res.QualityReason = model.ReasonOutOfWarningRange
			addFlag(model.FlagRangeWarning)
		}

		// Warning Maximum (Soft Limit)
		if res.Quality != model.QualityBad && param.WarningHigh != nil && in.Value > *param.WarningHigh {
			res.Quality = model.QualityUncertain
			res.QualityReason = model.ReasonOutOfWarningRange
			addFlag(model.FlagRangeWarning)
		}
	}

	// 6. Spike Detection (Deterministic rolling delta check)
	if param != nil && param.SpikeDetectionEnabled && param.SpikeThreshold > 0 && res.Quality != model.QualityBad {
		windowSize := param.SpikeWindowSize
		if windowSize <= 0 {
			windowSize = 3
		}

		if len(hist.points) > 0 {
			lastPt := hist.points[len(hist.points)-1]
			delta := math.Abs(in.Value - lastPt.Value)
			if delta > param.SpikeThreshold {
				res.IsSpike = true
				res.Quality = model.QualityUncertain
				res.QualityReason = model.ReasonSpikeDetected
				addFlag(model.FlagSpike)
			}
		}
	}

	// 7. Anomaly Debounce Buffer Hold State
	if in.IsHeldValue && res.Quality == model.QualityGood {
		res.Quality = model.QualityUncertain
		res.QualityReason = model.ReasonHoldAnomalyGrace
		addFlag(model.FlagHeldAnomaly)
	}

	// 8. Update Bounded In-Memory History (Max 10 points)
	maxPoints := 10
	if param != nil && param.SpikeWindowSize > maxPoints {
		maxPoints = param.SpikeWindowSize
	}
	if len(hist.points) >= maxPoints {
		hist.points = hist.points[1:]
	}
	hist.points = append(hist.points, historyPoint{
		Timestamp: recAt,
		Value:     in.Value,
		RawValue:  in.RawValue,
	})

	if res.Quality == model.QualityGood {
		hist.lastValidTime = recAt
	}
	hist.lastRecTime = recAt
	hist.lastValue = in.Value

	res.QualityFlags = strings.Join(flags, ",")
	return res
}

// CheckStale determines if a parameter's data has become stale based on elapsed time
func (qp *QualityProcessor) CheckStale(paramID uint, staleTimeoutSec int, lastDataTime time.Time) (bool, model.TelemetryQuality, model.QualityReason) {
	if staleTimeoutSec <= 0 {
		staleTimeoutSec = 120 // Default 2 minutes
	}

	now := time.Now().UTC()
	hist := qp.getOrCreateHistory(paramID)
	hist.mu.RLock()
	refTime := hist.lastValidTime
	if refTime.IsZero() {
		refTime = hist.lastRecTime
	}
	hist.mu.RUnlock()

	if refTime.IsZero() {
		refTime = lastDataTime
	}

	if refTime.IsZero() {
		return false, model.QualityUnknown, model.ReasonNone
	}

	elapsed := now.Sub(refTime.UTC())
	if elapsed > time.Duration(staleTimeoutSec)*time.Second {
		return true, model.QualityStale, model.ReasonStaleData
	}

	return false, model.QualityGood, model.ReasonNone
}

// FormatQualitySummary produces formatted statistics for diagnostic monitoring
func FormatQualitySummary(good, uncertain, bad, stale int) string {
	total := good + uncertain + bad + stale
	if total == 0 {
		return "No active parameters"
	}
	goodPct := float64(good) / float64(total) * 100.0
	return fmt.Sprintf("Health: %.1f%% Good (%d Good, %d Uncertain, %d Bad, %d Stale)",
		goodPct, good, uncertain, bad, stale)
}

// Reset clears history buffer for a specific parameter (useful on device reload or config change)
func (qp *QualityProcessor) Reset(paramID uint) {
	qp.historyMu.Lock()
	defer qp.historyMu.Unlock()
	delete(qp.history, paramID)
}

// ResetAll clears all parameter histories (e.g. on service restart)
func (qp *QualityProcessor) ResetAll() {
	qp.historyMu.Lock()
	defer qp.historyMu.Unlock()
	qp.history = make(map[uint]*paramHistory)
}

