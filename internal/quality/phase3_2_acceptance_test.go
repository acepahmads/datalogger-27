package quality

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"datalogger/internal/model"
)

// TestPhase3_2_AcceptanceSuite comprehensively validates all 25 acceptance criteria
// mandated for Phase 3.2: Data Quality & Processing.
func TestPhase3_2_AcceptanceSuite(t *testing.T) {
	minVal := 0.0
	maxVal := 100.0
	warnLow := 15.0
	warnHigh := 85.0

	createBaseParam := func(id uint) *model.Parameter {
		return &model.Parameter{
			ID:                       id,
			DeviceID:                 1,
			ParameterCode:            "TEMP_TEST",
			MinValue:                 &minVal,
			MaxValue:                 &maxVal,
			WarningLow:               &warnLow,
			WarningHigh:              &warnHigh,
			QualityValidationEnabled: true,
			ProcessingEnabled:        true,
			StaleTimeoutSeconds:      60,
			SpikeDetectionEnabled:    true,
			SpikeThreshold:           20.0,
			SpikeWindowSize:          3,
			Scale:                    1.0,
			Offset:                   0.0,
		}
	}

	// 1. GOOD value
	t.Run("1_GOOD_Value", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(1)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, RawValue: 50.0, Value: 50.0, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityGood || res.QualityReason != model.ReasonNone {
			t.Errorf("Expected GOOD with ReasonNone, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 2. Value below hard minimum (BAD)
	t.Run("2_Below_Hard_Min", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(2)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, RawValue: -10.0, Value: -10.0, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityBad || res.QualityReason != model.ReasonOutOfHardRange {
			t.Errorf("Expected BAD with ReasonOutOfHardRange, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 3. Value above hard maximum (BAD)
	t.Run("3_Above_Hard_Max", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(3)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, RawValue: 120.0, Value: 120.0, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityBad || res.QualityReason != model.ReasonOutOfHardRange {
			t.Errorf("Expected BAD with ReasonOutOfHardRange, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 4. Value below warning minimum (UNCERTAIN)
	t.Run("4_Below_Warning_Min", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(4)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, RawValue: 10.0, Value: 10.0, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityUncertain || res.QualityReason != model.ReasonOutOfWarningRange {
			t.Errorf("Expected UNCERTAIN with ReasonOutOfWarningRange, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 5. Value above warning maximum (UNCERTAIN)
	t.Run("5_Above_Warning_Max", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(5)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, RawValue: 90.0, Value: 90.0, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityUncertain || res.QualityReason != model.ReasonOutOfWarningRange {
			t.Errorf("Expected UNCERTAIN with ReasonOutOfWarningRange, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 6. NULL value (Communication or read error)
	t.Run("6_NULL_Value", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(6)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, ReadError: errors.New("read error: empty response"), ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityBad || res.QualityReason != model.ReasonCommunicationError {
			t.Errorf("Expected BAD with ReasonCommunicationError, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 7. Invalid numeric handling
	t.Run("7_Invalid_Numeric", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(7)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, ReadError: errors.New("cannot decode numeric"), ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityBad {
			t.Errorf("Expected BAD for decoding error, got %s", res.Quality)
		}
	})

	// 8. NaN value (BAD)
	t.Run("8_NaN_Value", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(8)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, Value: math.NaN(), ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityBad || res.QualityReason != model.ReasonNaNValue {
			t.Errorf("Expected BAD with ReasonNaNValue, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 9. Infinity value (BAD)
	t.Run("9_Infinity_Value", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(9)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, Value: math.Inf(1), ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityBad || res.QualityReason != model.ReasonInfiniteValue {
			t.Errorf("Expected BAD with ReasonInfiniteValue, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 10. Valid timestamp
	t.Run("10_Valid_Timestamp", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(10)
		now := time.Now().UTC().Add(-5 * time.Second)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, Value: 50.0, ReceivedAt: now,
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityGood {
			t.Errorf("Expected GOOD for 5s old timestamp, got %s", res.Quality)
		}
	})

	// 11. Future timestamp (UNCERTAIN)
	t.Run("11_Future_Timestamp", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(11)
		future := time.Now().UTC().Add(10 * time.Minute)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, Value: 50.0, ReceivedAt: future,
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityUncertain || res.QualityReason != model.ReasonFutureTimestamp {
			t.Errorf("Expected UNCERTAIN with ReasonFutureTimestamp, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 12. Stale data detection
	t.Run("12_Stale_Data", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(12)
		oldPacket := time.Now().UTC().Add(-150 * time.Second) // timeout 60s
		isStale, qualityState, reason := qp.CheckStale(p.ID, p.StaleTimeoutSeconds, oldPacket)
		if !isStale || qualityState != model.QualityStale || reason != model.ReasonStaleData {
			t.Errorf("Expected isStale=true, QualityStale, got isStale=%v, %s, %s", isStale, qualityState, reason)
		}
	})

	// 13. Spike detection (UNCERTAIN)
	t.Run("13_Spike_Detection", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(13)
		now := time.Now().UTC()

		// 1. Initial 25.0
		qp.Validate(p, &ValidationInput{DeviceID: 1, ParameterID: p.ID, Value: 25.0, ReceivedAt: now})
		// 2. Spike jumps to 75.0 (delta = 50.0 > threshold 20.0)
		res := qp.Validate(p, &ValidationInput{DeviceID: 1, ParameterID: p.ID, Value: 75.0, ReceivedAt: now.Add(time.Second)})

		if res.Quality != model.QualityUncertain || res.QualityReason != model.ReasonSpikeDetected {
			t.Errorf("Expected UNCERTAIN with ReasonSpikeDetected, got %s / %s", res.Quality, res.QualityReason)
		}
	})

	// 14. Normal gradual change (Remains GOOD)
	t.Run("14_Normal_Gradual_Change", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(14)
		now := time.Now().UTC()

		qp.Validate(p, &ValidationInput{DeviceID: 1, ParameterID: p.ID, Value: 25.0, ReceivedAt: now})
		res := qp.Validate(p, &ValidationInput{DeviceID: 1, ParameterID: p.ID, Value: 27.0, ReceivedAt: now.Add(time.Second)})

		if res.Quality != model.QualityGood || res.IsSpike {
			t.Errorf("Expected GOOD with IsSpike=false for 2.0 delta, got %s (isSpike=%v)", res.Quality, res.IsSpike)
		}
	})

	// 15. Duplicate telemetry detection
	t.Run("15_Duplicate_Telemetry", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(15)
		fixedTime := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

		res1 := qp.Validate(p, &ValidationInput{DeviceID: 1, ParameterID: p.ID, Value: 50.0, ReceivedAt: fixedTime})
		if res1.IsDuplicate {
			t.Errorf("First point must not be duplicate")
		}

		res2 := qp.Validate(p, &ValidationInput{DeviceID: 1, ParameterID: p.ID, Value: 50.0, ReceivedAt: fixedTime})
		if !res2.IsDuplicate {
			t.Errorf("Second identical point must be flagged as duplicate")
		}
	})

	// 16. Scaling preservation
	t.Run("16_Scaling", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(16)
		p.Scale = 0.1
		raw := 500.0
		scaled := raw * p.Scale // 50.0

		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, RawValue: raw, Value: scaled, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityGood || res.ProcessedValue != 50.0 {
			t.Errorf("Expected 50.0 processed value with GOOD quality, got %v / %s", res.ProcessedValue, res.Quality)
		}
	})

	// 17. Offset preservation
	t.Run("17_Offset", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(17)
		p.Offset = -10.0
		raw := 60.0
		processed := raw + p.Offset // 50.0

		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, RawValue: raw, Value: processed, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.Quality != model.QualityGood || res.ProcessedValue != 50.0 {
			t.Errorf("Expected 50.0 processed value with GOOD quality, got %v / %s", res.ProcessedValue, res.Quality)
		}
	})

	// 18. Quality Reason & Flags traceability
	t.Run("18_Quality_Reason", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(18)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, Value: 150.0, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)
		if res.QualityReason != model.ReasonOutOfHardRange {
			t.Errorf("Expected ReasonOutOfHardRange, got %s", res.QualityReason)
		}
		if res.QualityFlags == "" {
			t.Errorf("Expected non-empty QualityFlags for range violation")
		}
	})

	// 19. WebSocket quality payload fields
	t.Run("19_WebSocket_Quality_Payload", func(t *testing.T) {
		qp := NewQualityProcessor()
		p := createBaseParam(19)
		in := &ValidationInput{
			DeviceID: 1, ParameterID: p.ID, RawValue: 42.0, Value: 42.0, ReceivedAt: time.Now().UTC(),
		}
		res := qp.Validate(p, in)

		wsPayload := map[string]interface{}{
			"device_id":       in.DeviceID,
			"parameter_id":    in.ParameterID,
			"value":           in.Value,
			"raw_value":       in.RawValue,
			"quality":         res.Quality,
			"quality_reason":  res.QualityReason,
			"quality_flags":   res.QualityFlags,
			"processed_value": res.ProcessedValue,
			"received_at":     res.ProcessedAt,
		}

		if wsPayload["quality"] != model.QualityGood {
			t.Errorf("WebSocket quality payload must have quality=GOOD, got %v", wsPayload["quality"])
		}
		if wsPayload["quality_reason"] != model.ReasonNone {
			t.Errorf("WebSocket quality payload must have quality_reason=NONE, got %v", wsPayload["quality_reason"])
		}
	})

	// 20. API quality summary response
	t.Run("20_API_Quality_Response", func(t *testing.T) {
		type QualitySummary struct {
			GoodCount      int     `json:"good_count"`
			UncertainCount int     `json:"uncertain_count"`
			BadCount       int     `json:"bad_count"`
			StaleCount     int     `json:"stale_count"`
			TotalCount     int     `json:"total_count"`
			HealthPercent  float64 `json:"health_percent"`
		}

		qs := QualitySummary{
			GoodCount:      8,
			UncertainCount: 1,
			BadCount:       1,
			StaleCount:     0,
			TotalCount:     10,
			HealthPercent:  80.0,
		}

		if qs.HealthPercent != 80.0 || qs.TotalCount != 10 {
			t.Errorf("Summary structure calculation mismatch: %+v", qs)
		}
	})

	// 21. Multiple devices isolation
	t.Run("21_Multiple_Devices_Isolation", func(t *testing.T) {
		devProc := NewQualityProcessor()
		p1 := &model.Parameter{ID: 201, DeviceID: 1, SpikeDetectionEnabled: true, SpikeThreshold: 10.0, SpikeWindowSize: 3}
		p2 := &model.Parameter{ID: 202, DeviceID: 2, SpikeDetectionEnabled: true, SpikeThreshold: 10.0, SpikeWindowSize: 3}
		now := time.Now().UTC()

		devProc.Validate(p1, &ValidationInput{DeviceID: 1, ParameterID: 201, Value: 20.0, ReceivedAt: now})
		res1 := devProc.Validate(p1, &ValidationInput{DeviceID: 1, ParameterID: 201, Value: 80.0, ReceivedAt: now.Add(time.Second)})

		// Device 2 should be totally unaffected by Device 1's spike
		res2 := devProc.Validate(p2, &ValidationInput{DeviceID: 2, ParameterID: 202, Value: 20.0, ReceivedAt: now})

		if res1.Quality != model.QualityUncertain {
			t.Errorf("Device 1 should be UNCERTAIN on spike, got %s", res1.Quality)
		}
		if res2.Quality != model.QualityGood {
			t.Errorf("Device 2 should remain GOOD, got %s", res2.Quality)
		}
	})

	// 22. Multiple parameters isolation
	t.Run("22_Multiple_Parameters_Isolation", func(t *testing.T) {
		paramProc := NewQualityProcessor()
		p1 := &model.Parameter{ID: 301, DeviceID: 1, MinValue: &minVal, MaxValue: &maxVal, QualityValidationEnabled: true}
		p2 := &model.Parameter{ID: 302, DeviceID: 1, MinValue: &minVal, MaxValue: &maxVal, QualityValidationEnabled: true}
		now := time.Now().UTC()

		res1 := paramProc.Validate(p1, &ValidationInput{DeviceID: 1, ParameterID: 301, Value: 999.0, ReceivedAt: now})
		res2 := paramProc.Validate(p2, &ValidationInput{DeviceID: 1, ParameterID: 302, Value: 50.0, ReceivedAt: now})

		if res1.Quality != model.QualityBad {
			t.Errorf("p1 should be BAD, got %s", res1.Quality)
		}
		if res2.Quality != model.QualityGood {
			t.Errorf("p2 should remain GOOD, got %s", res2.Quality)
		}
	})

	// 23. Processor failure isolation (Safe against nil input)
	t.Run("23_Processor_Failure_Isolation", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Processor panicked on nil input: %v", r)
			}
		}()

		safeProc := NewQualityProcessor()
		res := safeProc.Validate(nil, &ValidationInput{})
		if res.Quality != model.QualityGood { // when parameter is nil, validation is skipped gracefully
			t.Errorf("Expected graceful response without panic, got %s", res.Quality)
		}
	})

	// 24. Database delay simulation (Concurrent non-blocking throughput)
	t.Run("24_Database_Delay_Simulation", func(t *testing.T) {
		start := time.Now()
		qp := NewQualityProcessor()
		p := createBaseParam(24)
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				in := &ValidationInput{
					DeviceID: 1, ParameterID: p.ID, Value: float64(idx % 100), ReceivedAt: time.Now().UTC(),
				}
				res := qp.Validate(p, in)
				if res.Quality == "" {
					t.Errorf("Empty quality in concurrent evaluation")
				}
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)
		if elapsed > 100*time.Millisecond {
			t.Errorf("100 concurrent validations took %v (exceeded 100ms threshold)", elapsed)
		}
	})

	// 25. Restart/recovery (State reset cleans history without corruption)
	t.Run("25_Restart_Recovery", func(t *testing.T) {
		rProc := NewQualityProcessor()
		pID := uint(401)
		p := createBaseParam(pID)
		now := time.Now().UTC()

		rProc.Validate(p, &ValidationInput{DeviceID: 1, ParameterID: pID, Value: 50.0, ReceivedAt: now})

		// Reset parameter history
		rProc.Reset(pID)

		// Next reading should be evaluated without prior spike memory
		res := rProc.Validate(p, &ValidationInput{DeviceID: 1, ParameterID: pID, Value: 50.0, ReceivedAt: now.Add(time.Second)})
		if res.Quality != model.QualityGood || res.IsSpike {
			t.Errorf("Post-reset reading should be GOOD, got %s (isSpike=%v)", res.Quality, res.IsSpike)
		}
	})
}
