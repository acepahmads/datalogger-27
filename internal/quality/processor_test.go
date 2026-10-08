package quality

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"datalogger/internal/model"
)

func TestQualityProcessor_GoodValue(t *testing.T) {
	qp := NewQualityProcessor()

	minVal := 0.0
	maxVal := 100.0
	warnLow := 10.0
	warnHigh := 90.0

	param := &model.Parameter{
		ID:                       1,
		ParameterCode:            "TEMP_CABINET",
		QualityValidationEnabled: true,
		MinValue:                 &minVal,
		MaxValue:                 &maxVal,
		WarningLow:               &warnLow,
		WarningHigh:              &warnHigh,
	}

	in := &ValidationInput{
		DeviceID:    1,
		ParameterID: 1,
		RawValue:    25.5,
		Value:       25.5,
		ReceivedAt:  time.Now().UTC(),
	}

	res := qp.Validate(param, in)
	if res.Quality != model.QualityGood {
		t.Errorf("expected QualityGood, got %v", res.Quality)
	}
	if res.QualityReason != model.ReasonNone {
		t.Errorf("expected ReasonNone, got %v", res.QualityReason)
	}
	if res.ProcessedValue != 25.5 {
		t.Errorf("expected ProcessedValue 25.5, got %v", res.ProcessedValue)
	}
}

func TestQualityProcessor_HardLimits(t *testing.T) {
	qp := NewQualityProcessor()

	minVal := 0.0
	maxVal := 100.0

	param := &model.Parameter{
		ID:                       2,
		ParameterCode:            "PRESSURE",
		QualityValidationEnabled: true,
		MinValue:                 &minVal,
		MaxValue:                 &maxVal,
	}

	// Below hard minimum
	inBelow := &ValidationInput{
		DeviceID:    1,
		ParameterID: 2,
		RawValue:    -5.0,
		Value:       -5.0,
		ReceivedAt:  time.Now().UTC(),
	}
	resBelow := qp.Validate(param, inBelow)
	if resBelow.Quality != model.QualityBad {
		t.Errorf("expected QualityBad for below hard min, got %v", resBelow.Quality)
	}
	if resBelow.QualityReason != model.ReasonOutOfHardRange {
		t.Errorf("expected ReasonOutOfHardRange, got %v", resBelow.QualityReason)
	}
	if !strings.Contains(resBelow.QualityFlags, model.FlagRangeViolation) {
		t.Errorf("expected flag range_violation, got %v", resBelow.QualityFlags)
	}

	// Above hard maximum
	inAbove := &ValidationInput{
		DeviceID:    1,
		ParameterID: 2,
		RawValue:    150.0,
		Value:       150.0,
		ReceivedAt:  time.Now().UTC(),
	}
	resAbove := qp.Validate(param, inAbove)
	if resAbove.Quality != model.QualityBad {
		t.Errorf("expected QualityBad for above hard max, got %v", resAbove.Quality)
	}
	if resAbove.QualityReason != model.ReasonOutOfHardRange {
		t.Errorf("expected ReasonOutOfHardRange, got %v", resAbove.QualityReason)
	}
}

func TestQualityProcessor_WarningLimits(t *testing.T) {
	qp := NewQualityProcessor()

	minVal := 0.0
	maxVal := 100.0
	warnLow := 15.0
	warnHigh := 85.0

	param := &model.Parameter{
		ID:                       3,
		ParameterCode:            "HUMIDITY",
		QualityValidationEnabled: true,
		MinValue:                 &minVal,
		MaxValue:                 &maxVal,
		WarningLow:               &warnLow,
		WarningHigh:              &warnHigh,
	}

	// In soft warning range (low warning)
	inLowWarn := &ValidationInput{
		DeviceID:    1,
		ParameterID: 3,
		RawValue:    12.0,
		Value:       12.0,
		ReceivedAt:  time.Now().UTC(),
	}
	resLowWarn := qp.Validate(param, inLowWarn)
	if resLowWarn.Quality != model.QualityUncertain {
		t.Errorf("expected QualityUncertain for warning low, got %v", resLowWarn.Quality)
	}
	if resLowWarn.QualityReason != model.ReasonOutOfWarningRange {
		t.Errorf("expected ReasonOutOfWarningRange, got %v", resLowWarn.QualityReason)
	}
	if !strings.Contains(resLowWarn.QualityFlags, model.FlagRangeWarning) {
		t.Errorf("expected flag range_warning, got %v", resLowWarn.QualityFlags)
	}

	// In soft warning range (high warning)
	inHighWarn := &ValidationInput{
		DeviceID:    1,
		ParameterID: 3,
		RawValue:    88.0,
		Value:       88.0,
		ReceivedAt:  time.Now().UTC(),
	}
	resHighWarn := qp.Validate(param, inHighWarn)
	if resHighWarn.Quality != model.QualityUncertain {
		t.Errorf("expected QualityUncertain for warning high, got %v", resHighWarn.Quality)
	}
	if resHighWarn.QualityReason != model.ReasonOutOfWarningRange {
		t.Errorf("expected ReasonOutOfWarningRange, got %v", resHighWarn.QualityReason)
	}
}

func TestQualityProcessor_NaNAndInfinity(t *testing.T) {
	qp := NewQualityProcessor()

	param := &model.Parameter{
		ID:                       4,
		ParameterCode:            "VOLTAGE",
		QualityValidationEnabled: true,
	}

	// NaN value
	inNaN := &ValidationInput{
		DeviceID:    1,
		ParameterID: 4,
		Value:       math.NaN(),
		ReceivedAt:  time.Now().UTC(),
	}
	resNaN := qp.Validate(param, inNaN)
	if resNaN.Quality != model.QualityBad {
		t.Errorf("expected QualityBad for NaN, got %v", resNaN.Quality)
	}
	if resNaN.QualityReason != model.ReasonNaNValue {
		t.Errorf("expected ReasonNaNValue, got %v", resNaN.QualityReason)
	}

	// Infinity value
	inInf := &ValidationInput{
		DeviceID:    1,
		ParameterID: 4,
		Value:       math.Inf(1),
		ReceivedAt:  time.Now().UTC(),
	}
	resInf := qp.Validate(param, inInf)
	if resInf.Quality != model.QualityBad {
		t.Errorf("expected QualityBad for Infinity, got %v", resInf.Quality)
	}
	if resInf.QualityReason != model.ReasonInfiniteValue {
		t.Errorf("expected ReasonInfiniteValue, got %v", resInf.QualityReason)
	}
}

func TestQualityProcessor_TimestampValidation(t *testing.T) {
	qp := NewQualityProcessor()
	param := &model.Parameter{
		ID:                       5,
		ParameterCode:            "FLOW",
		QualityValidationEnabled: true,
	}

	// Future timestamp (+5 minutes)
	inFuture := &ValidationInput{
		DeviceID:    1,
		ParameterID: 5,
		Value:       10.0,
		ReceivedAt:  time.Now().UTC().Add(5 * time.Minute),
	}
	resFuture := qp.Validate(param, inFuture)
	if resFuture.Quality != model.QualityUncertain {
		t.Errorf("expected QualityUncertain for future timestamp, got %v", resFuture.Quality)
	}
	if resFuture.QualityReason != model.ReasonFutureTimestamp {
		t.Errorf("expected ReasonFutureTimestamp, got %v", resFuture.QualityReason)
	}
	if !strings.Contains(resFuture.QualityFlags, model.FlagTimestampWarn) {
		t.Errorf("expected flag timestamp_warning, got %v", resFuture.QualityFlags)
	}
}

func TestQualityProcessor_SpikeDetection(t *testing.T) {
	qp := NewQualityProcessor()
	param := &model.Parameter{
		ID:                       6,
		ParameterCode:            "TEMPERATURE",
		QualityValidationEnabled: true,
		SpikeDetectionEnabled:    true,
		SpikeThreshold:           10.0, // Max delta 10°C allowed between reads
		SpikeWindowSize:          3,
	}

	now := time.Now().UTC()

	// 1. Initial normal read: 25°C
	res1 := qp.Validate(param, &ValidationInput{
		DeviceID: 1, ParameterID: 6, Value: 25.0, ReceivedAt: now,
	})
	if res1.Quality != model.QualityGood || res1.IsSpike {
		t.Fatalf("first reading should be normal Good, got %v", res1.Quality)
	}

	// 2. Normal gradual change: 27°C (delta = 2.0 <= 10.0)
	res2 := qp.Validate(param, &ValidationInput{
		DeviceID: 1, ParameterID: 6, Value: 27.0, ReceivedAt: now.Add(1 * time.Second),
	})
	if res2.Quality != model.QualityGood || res2.IsSpike {
		t.Fatalf("gradual reading should be Good, got %v (isSpike=%v)", res2.Quality, res2.IsSpike)
	}

	// 3. Sudden spike: 65°C (delta = 38.0 > 10.0!)
	res3 := qp.Validate(param, &ValidationInput{
		DeviceID: 1, ParameterID: 6, Value: 65.0, ReceivedAt: now.Add(2 * time.Second),
	})
	if res3.Quality != model.QualityUncertain {
		t.Errorf("expected QualityUncertain on spike, got %v", res3.Quality)
	}
	if res3.QualityReason != model.ReasonSpikeDetected {
		t.Errorf("expected ReasonSpikeDetected, got %v", res3.QualityReason)
	}
	if !res3.IsSpike {
		t.Errorf("expected IsSpike=true")
	}
}

func TestQualityProcessor_DuplicateDetection(t *testing.T) {
	qp := NewQualityProcessor()
	param := &model.Parameter{
		ID:                       7,
		ParameterCode:            "LEVEL",
		QualityValidationEnabled: true,
	}

	ts := time.Now().UTC().Truncate(time.Second)

	// First reading
	res1 := qp.Validate(param, &ValidationInput{
		DeviceID: 1, ParameterID: 7, Value: 50.0, ReceivedAt: ts,
	})
	if res1.IsDuplicate {
		t.Errorf("initial reading should not be duplicate")
	}

	// Exact duplicate reading (same param, same timestamp, same value)
	res2 := qp.Validate(param, &ValidationInput{
		DeviceID: 1, ParameterID: 7, Value: 50.0, ReceivedAt: ts,
	})
	if !res2.IsDuplicate {
		t.Errorf("expected duplicate to be flagged")
	}
	if !strings.Contains(res2.QualityFlags, model.FlagDuplicate) {
		t.Errorf("expected flag duplicate, got %v", res2.QualityFlags)
	}
}

func TestQualityProcessor_StaleDetection(t *testing.T) {
	qp := NewQualityProcessor()

	// 1. Fresh reading
	recentTime := time.Now().UTC().Add(-10 * time.Second)
	isStale, _, _ := qp.CheckStale(8, 60, recentTime)
	if isStale {
		t.Errorf("data from 10s ago should not be stale with 60s timeout")
	}

	// 2. Stale reading (older than 60s timeout)
	staleTime := time.Now().UTC().Add(-125 * time.Second)
	isStale, quality, reason := qp.CheckStale(8, 60, staleTime)
	if !isStale {
		t.Errorf("expected data from 125s ago to be stale with 60s timeout")
	}
	if quality != model.QualityStale {
		t.Errorf("expected QualityStale, got %v", quality)
	}
	if reason != model.ReasonStaleData {
		t.Errorf("expected ReasonStaleData, got %v", reason)
	}
}

func TestQualityProcessor_HeldAnomalyGrace(t *testing.T) {
	qp := NewQualityProcessor()
	param := &model.Parameter{
		ID:                       9,
		ParameterCode:            "ENERGY",
		QualityValidationEnabled: true,
	}

	inHeld := &ValidationInput{
		DeviceID:    1,
		ParameterID: 9,
		Value:       123.4,
		IsHeldValue: true, // Marked as held value by debounce tracker
		ReceivedAt:  time.Now().UTC(),
	}
	resHeld := qp.Validate(param, inHeld)
	if resHeld.Quality != model.QualityUncertain {
		t.Errorf("expected QualityUncertain for held value, got %v", resHeld.Quality)
	}
	if resHeld.QualityReason != model.ReasonHoldAnomalyGrace {
		t.Errorf("expected ReasonHoldAnomalyGrace, got %v", resHeld.QualityReason)
	}
}

func TestQualityProcessor_CommunicationError(t *testing.T) {
	qp := NewQualityProcessor()
	param := &model.Parameter{
		ID:                       10,
		ParameterCode:            "CURRENT",
		QualityValidationEnabled: true,
	}

	inErr := &ValidationInput{
		DeviceID:    1,
		ParameterID: 10,
		ReadError:   errors.New("i/o timeout on Modbus RTU"),
		ReceivedAt:  time.Now().UTC(),
	}
	resErr := qp.Validate(param, inErr)
	if resErr.Quality != model.QualityBad {
		t.Errorf("expected QualityBad for read error, got %v", resErr.Quality)
	}
	if resErr.QualityReason != model.ReasonCommunicationError {
		t.Errorf("expected ReasonCommunicationError, got %v", resErr.QualityReason)
	}
}
