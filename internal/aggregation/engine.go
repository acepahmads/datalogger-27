package aggregation

import (
	"math"
	"sort"
	"time"

	"datalogger/internal/model"
)

// CalculationSample represents an extracted, validated data point for aggregation
type CalculationSample struct {
	Timestamp time.Time
	Value     float64
	Quality   model.TelemetryQuality
	Reason    model.QualityReason
	IsValid   bool
}

// Engine performs deterministic statistical and quality rollups over raw telemetry collections
type Engine struct{}

// NewEngine creates an aggregation calculation engine
func NewEngine() *Engine {
	return &Engine{}
}

// Calculate executes the aggregation definition across the provided telemetry samples for a time bucket
func (e *Engine) Calculate(
	def *model.AggregationDefinition,
	bucketStart time.Time,
	bucketEnd time.Time,
	samples []model.RawData,
) *model.AggregationResult {
	loc := ResolveLocation(def.Timezone)
	identifier := FormatIdentifier(bucketEnd, loc)

	res := &model.AggregationResult{
		AggregationDefinitionID: def.ID,
		DeviceID:                def.DeviceID,
		ParameterID:             def.ParameterID,
		SourceType:              def.SourceType,
		PeriodStart:             bucketStart,
		PeriodEnd:               bucketEnd,
		Identifier:              identifier,
		Function:                def.Function,
		IntervalSeconds:         def.IntervalSeconds,
		SampleCount:             len(samples),
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}

	if len(samples) == 0 {
		res.Quality = model.QualityBad
		res.QualityReason = model.ReasonNullValue
		return res
	}

// MaxPlausibleEnvironmentalValue defines the physical upper bound for environmental telemetry (100,000,000 µg/m³).
// Any value exceeding this magnitude represents a hardware register overflow / corrupted transmission (e.g. 3.69e19)
// and is automatically excluded from mathematical aggregation by the Smart Aggregation Sanitizer (POS 2).
const MaxPlausibleEnvironmentalValue = 1e8

	// 1. Extract values based on source domain (INTERNAL_RAW vs CUSTOMER_PROCESSED)
	isRaw := def.SourceType == model.SourceInternalRaw
	var validSamples []CalculationSample

	for _, s := range samples {
		var val float64
		if isRaw {
			val = s.RawValue
		} else {
			// Customer Processed: Uses Phase 3.2 processed value
			val = s.ProcessedValue
		}

		// Validate numeric viability (guard against NaN, Inf, and unphysical register overflow)
		isCorrupt := math.IsNaN(val) || math.IsInf(val, 0) || math.Abs(val) > MaxPlausibleEnvironmentalValue

		// Quality counting
		if isCorrupt {
			res.BadCount++
		} else {
			switch s.Quality {
			case model.QualityGood:
				res.GoodCount++
			case model.QualityUncertain:
				res.UncertainCount++
			case model.QualityBad:
				res.BadCount++
			case model.QualityStale:
				res.StaleCount++
			default:
				res.UncertainCount++
			}
		}

		if isCorrupt {
			continue
		}

		// Check timestamp priority
		ts := s.ReceivedAt
		if s.DeviceTimestamp != nil && !s.DeviceTimestamp.IsZero() {
			ts = *s.DeviceTimestamp
		}

		validSamples = append(validSamples, CalculationSample{
			Timestamp: ts,
			Value:     val,
			Quality:   s.Quality,
			Reason:    s.QualityReason,
			IsValid:   true,
		})
	}

	res.ValidCount = len(validSamples)

	// If zero valid samples, return with BAD quality and nil value
	if len(validSamples) == 0 {
		res.Quality = model.QualityBad
		res.QualityReason = model.ReasonNullValue
		return res
	}

	// Sort samples by timestamp for deterministic FIRST and LAST calculation
	sort.SliceStable(validSamples, func(i, j int) bool {
		return validSamples[i].Timestamp.Before(validSamples[j].Timestamp)
	})

	// POS 2: Smart Aggregation Sample Sanitization
	// If healthy (GOOD / UNCERTAIN) samples exist, exclude QualityBad samples from mathematical calculation
	// so faulty registers or sensor read errors do not corrupt the bucket value or average.
	var calcSamples []CalculationSample
	for _, s := range validSamples {
		if s.Quality != model.QualityBad {
			calcSamples = append(calcSamples, s)
		}
	}
	if len(calcSamples) == 0 {
		calcSamples = validSamples
	}

	// 2. Calculate Statistical Metrics on Sanitized Samples
	sum := 0.0
	minVal := calcSamples[0].Value
	maxVal := calcSamples[0].Value
	firstVal := calcSamples[0].Value
	lastVal := calcSamples[len(calcSamples)-1].Value

	for _, s := range calcSamples {
		sum += s.Value
		if s.Value < minVal {
			minVal = s.Value
		}
		if s.Value > maxVal {
			maxVal = s.Value
		}
	}

	avgVal := sum / float64(len(calcSamples))

	res.MinValue = &minVal
	res.MaxValue = &maxVal
	res.AvgValue = &avgVal
	res.SumValue = &sum
	res.FirstValue = &firstVal
	res.LastValue = &lastVal

	// 3. Assign Primary Result Value based on configured function
	switch def.Function {
	case model.FunctionMin:
		v := minVal
		res.Value = &v
	case model.FunctionMax:
		v := maxVal
		res.Value = &v
	case model.FunctionSum:
		v := sum
		res.Value = &v
	case model.FunctionCount:
		v := float64(len(validSamples))
		res.Value = &v
	case model.FunctionFirst:
		v := firstVal
		res.Value = &v
	case model.FunctionLast:
		v := lastVal
		res.Value = &v
	case model.FunctionAvg:
		fallthrough
	default:
		v := avgVal
		res.Value = &v
	}

	// 4. Determine Overall Aggregated Quality
	// Rule:
	// - If any BAD sample exists: quality = UNCERTAIN (or BAD if all samples bad)
	// - If any UNCERTAIN sample exists: quality = UNCERTAIN
	// - If any STALE sample exists: quality = STALE (if no bad/uncertain)
	// - If all samples are GOOD: quality = GOOD
	if res.BadCount > 0 {
		if res.GoodCount == 0 && res.UncertainCount == 0 {
			res.Quality = model.QualityBad
			res.QualityReason = model.ReasonOutOfHardRange
		} else {
			res.Quality = model.QualityUncertain
			res.QualityReason = model.ReasonOutOfWarningRange
		}
	} else if res.UncertainCount > 0 {
		res.Quality = model.QualityUncertain
		res.QualityReason = model.ReasonOutOfWarningRange
	} else if res.StaleCount > 0 {
		res.Quality = model.QualityStale
		res.QualityReason = model.ReasonStaleData
	} else {
		res.Quality = model.QualityGood
		res.QualityReason = model.ReasonNone
	}

	return res
}
