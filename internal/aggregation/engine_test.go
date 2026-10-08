package aggregation_test

import (
	"math"
	"testing"
	"time"

	"datalogger/internal/aggregation"
	"datalogger/internal/model"
)

func TestInternalRawAggregationFunctions(t *testing.T) {
	eng := aggregation.NewEngine()
	now := time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC)
	bEnd := now.Add(5 * time.Minute)

	samples := []model.RawData{
		{RawValue: 25.1, ProcessedValue: 100.0, ReceivedAt: now.Add(10 * time.Second), Quality: model.QualityGood},
		{RawValue: 25.2, ProcessedValue: 101.0, ReceivedAt: now.Add(20 * time.Second), Quality: model.QualityGood},
		{RawValue: 25.4, ProcessedValue: 102.0, ReceivedAt: now.Add(30 * time.Second), Quality: model.QualityGood},
		{RawValue: 25.3, ProcessedValue: 103.0, ReceivedAt: now.Add(40 * time.Second), Quality: model.QualityGood},
		{RawValue: 25.2, ProcessedValue: 104.0, ReceivedAt: now.Add(50 * time.Second), Quality: model.QualityGood},
	}

	baseDef := &model.AggregationDefinition{
		ID:              1,
		DeviceID:        10,
		ParameterID:     101,
		SourceType:      model.SourceInternalRaw,
		IntervalSeconds: 300,
		Timezone:        "UTC",
	}

	// 1. Internal Raw AVG
	baseDef.Function = model.FunctionAvg
	resAvg := eng.Calculate(baseDef, now, bEnd, samples)
	if resAvg.Value == nil || math.Abs(*resAvg.Value-25.24) > 0.001 {
		t.Errorf("Expected raw AVG 25.24, got %v", resAvg.Value)
	}

	// 2. Internal Raw MIN
	baseDef.Function = model.FunctionMin
	resMin := eng.Calculate(baseDef, now, bEnd, samples)
	if resMin.Value == nil || *resMin.Value != 25.1 {
		t.Errorf("Expected raw MIN 25.1, got %v", resMin.Value)
	}

	// 3. Internal Raw MAX
	baseDef.Function = model.FunctionMax
	resMax := eng.Calculate(baseDef, now, bEnd, samples)
	if resMax.Value == nil || *resMax.Value != 25.4 {
		t.Errorf("Expected raw MAX 25.4, got %v", resMax.Value)
	}

	// 4. Internal Raw SUM
	baseDef.Function = model.FunctionSum
	resSum := eng.Calculate(baseDef, now, bEnd, samples)
	if resSum.Value == nil || math.Abs(*resSum.Value-126.2) > 0.001 {
		t.Errorf("Expected raw SUM 126.2, got %v", resSum.Value)
	}

	// 5. Internal Raw COUNT
	baseDef.Function = model.FunctionCount
	resCount := eng.Calculate(baseDef, now, bEnd, samples)
	if resCount.Value == nil || *resCount.Value != 5 {
		t.Errorf("Expected raw COUNT 5, got %v", resCount.Value)
	}

	// 6. Internal Raw FIRST
	baseDef.Function = model.FunctionFirst
	resFirst := eng.Calculate(baseDef, now, bEnd, samples)
	if resFirst.Value == nil || *resFirst.Value != 25.1 {
		t.Errorf("Expected raw FIRST 25.1, got %v", resFirst.Value)
	}

	// 7. Internal Raw LAST
	baseDef.Function = model.FunctionLast
	resLast := eng.Calculate(baseDef, now, bEnd, samples)
	if resLast.Value == nil || *resLast.Value != 25.2 {
		t.Errorf("Expected raw LAST 25.2, got %v", resLast.Value)
	}
}

func TestCustomerProcessedAggregationFunctions(t *testing.T) {
	eng := aggregation.NewEngine()
	now := time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC)
	bEnd := now.Add(30 * time.Minute)

	// Raw: 100, 110, 120
	// Formula: (raw * 0.1) + 2 => Processed: 12, 13, 14
	samples := []model.RawData{
		{RawValue: 100.0, ProcessedValue: 12.0, ReceivedAt: now.Add(1 * time.Minute), Quality: model.QualityGood},
		{RawValue: 110.0, ProcessedValue: 13.0, ReceivedAt: now.Add(2 * time.Minute), Quality: model.QualityGood},
		{RawValue: 120.0, ProcessedValue: 14.0, ReceivedAt: now.Add(3 * time.Minute), Quality: model.QualityGood},
	}

	baseDef := &model.AggregationDefinition{
		ID:              2,
		DeviceID:        10,
		ParameterID:     102,
		SourceType:      model.SourceCustomerProcessed,
		IntervalSeconds: 1800,
		Timezone:        "UTC",
	}

	// 8. Customer Processed AVG (must be 13, NOT average of raw 100, 110, 120)
	baseDef.Function = model.FunctionAvg
	resAvg := eng.Calculate(baseDef, now, bEnd, samples)
	if resAvg.Value == nil || *resAvg.Value != 13.0 {
		t.Fatalf("Expected customer processed AVG 13.0, got %v", resAvg.Value)
	}

	// 9. Customer Processed MIN (must be 12.0)
	baseDef.Function = model.FunctionMin
	resMin := eng.Calculate(baseDef, now, bEnd, samples)
	if resMin.Value == nil || *resMin.Value != 12.0 {
		t.Errorf("Expected customer processed MIN 12.0, got %v", resMin.Value)
	}

	// 10. Customer Processed MAX (must be 14.0)
	baseDef.Function = model.FunctionMax
	resMax := eng.Calculate(baseDef, now, bEnd, samples)
	if resMax.Value == nil || *resMax.Value != 14.0 {
		t.Errorf("Expected customer processed MAX 14.0, got %v", resMax.Value)
	}

	// 11. Customer Processed SUM (must be 39.0)
	baseDef.Function = model.FunctionSum
	resSum := eng.Calculate(baseDef, now, bEnd, samples)
	if resSum.Value == nil || *resSum.Value != 39.0 {
		t.Errorf("Expected customer processed SUM 39.0, got %v", resSum.Value)
	}

	// 12. Customer Processed COUNT (must be 3)
	baseDef.Function = model.FunctionCount
	resCount := eng.Calculate(baseDef, now, bEnd, samples)
	if resCount.Value == nil || *resCount.Value != 3.0 {
		t.Errorf("Expected customer processed COUNT 3.0, got %v", resCount.Value)
	}
}

func TestConfigurableIntervalsAndTimeBuckets(t *testing.T) {
	testTime := time.Date(2026, 10, 8, 14, 17, 35, 0, time.UTC)

	// 13. 2-minute interval (120s) => 14:16:00 to 14:18:00
	b2m := aggregation.CalculateBucket(testTime, 120, "UTC")
	if b2m.Identifier != "20261008141600" {
		t.Errorf("Expected 2m bucket identifier 20261008141600, got %s", b2m.Identifier)
	}
	if b2m.PeriodEnd.Sub(b2m.PeriodStart) != 2*time.Minute {
		t.Errorf("Expected 2m duration, got %v", b2m.PeriodEnd.Sub(b2m.PeriodStart))
	}

	// 14. 5-minute interval (300s) => 14:15:00 to 14:20:00
	b5m := aggregation.CalculateBucket(testTime, 300, "UTC")
	if b5m.Identifier != "20261008141500" {
		t.Errorf("Expected 5m bucket identifier 20261008141500, got %s", b5m.Identifier)
	}

	// 15. 30-minute interval (1800s) => 14:00:00 to 14:30:00
	b30m := aggregation.CalculateBucket(testTime, 1800, "UTC")
	if b30m.Identifier != "20261008140000" {
		t.Errorf("Expected 30m bucket identifier 20261008140000, got %s", b30m.Identifier)
	}

	// 16. 60-minute interval (3600s) => 14:00:00 to 15:00:00
	b60m := aggregation.CalculateBucket(testTime, 3600, "UTC")
	if b60m.Identifier != "20261008140000" {
		t.Errorf("Expected 60m bucket identifier 20261008140000, got %s", b60m.Identifier)
	}

	// 17. Custom interval (15m = 900s) => 14:15:00 to 14:30:00
	b15m := aggregation.CalculateBucket(testTime, 900, "UTC")
	if b15m.Identifier != "20261008141500" {
		t.Errorf("Expected 15m bucket identifier 20261008141500, got %s", b15m.Identifier)
	}
}

func TestTimezoneAndIdentifierParsing(t *testing.T) {
	// 18 & 19 & 20: Deterministic Identifier & Timezone
	utcTime := time.Date(2026, 10, 8, 7, 0, 0, 0, time.UTC) // 07:00 UTC = 14:00 Asia/Jakarta (WIB)

	bWIB := aggregation.CalculateBucket(utcTime, 1800, "Asia/Jakarta")
	if bWIB.Identifier != "20261008140000" {
		t.Errorf("Expected Jakarta identifier 20261008140000, got %s", bWIB.Identifier)
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	parsed, err := aggregation.ParseIdentifier("20261008140000", loc)
	if err != nil {
		t.Fatalf("Failed to parse identifier: %v", err)
	}
	if parsed.UTC() != utcTime {
		t.Errorf("Expected parsed time in UTC %v, got %v", utcTime, parsed.UTC())
	}
}

func TestQualityAggregationAndInvalidSamples(t *testing.T) {
	eng := aggregation.NewEngine()
	now := time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC)
	bEnd := now.Add(5 * time.Minute)

	def := &model.AggregationDefinition{
		ID:              3,
		SourceType:      model.SourceCustomerProcessed,
		Function:        model.FunctionAvg,
		IntervalSeconds: 300,
		Timezone:        "UTC",
	}

	// 21. Quality accounting with mixed GOOD, UNCERTAIN, BAD, STALE
	mixedSamples := []model.RawData{
		{ProcessedValue: 20.0, ReceivedAt: now.Add(10 * time.Second), Quality: model.QualityGood},
		{ProcessedValue: 22.0, ReceivedAt: now.Add(20 * time.Second), Quality: model.QualityGood},
		{ProcessedValue: 24.0, ReceivedAt: now.Add(30 * time.Second), Quality: model.QualityUncertain},
		{ProcessedValue: 26.0, ReceivedAt: now.Add(40 * time.Second), Quality: model.QualityBad},
	}

	resMixed := eng.Calculate(def, now, bEnd, mixedSamples)
	if resMixed.SampleCount != 4 || resMixed.GoodCount != 2 || resMixed.UncertainCount != 1 || resMixed.BadCount != 1 {
		t.Errorf("Quality count mismatch: Sample=%d, Good=%d, Uncertain=%d, Bad=%d",
			resMixed.SampleCount, resMixed.GoodCount, resMixed.UncertainCount, resMixed.BadCount)
	}
	if resMixed.Quality != model.QualityUncertain {
		t.Errorf("Expected overall quality UNCERTAIN due to bad and good mix, got %s", resMixed.Quality)
	}

	// 22. Invalid Sample Handling (NaN, Inf handled without crash)
	nanSamples := []model.RawData{
		{ProcessedValue: 10.0, ReceivedAt: now.Add(10 * time.Second), Quality: model.QualityGood},
		{ProcessedValue: math.NaN(), ReceivedAt: now.Add(20 * time.Second), Quality: model.QualityBad},
		{ProcessedValue: math.Inf(1), ReceivedAt: now.Add(30 * time.Second), Quality: model.QualityBad},
		{ProcessedValue: 20.0, ReceivedAt: now.Add(40 * time.Second), Quality: model.QualityGood},
	}
	resNaN := eng.Calculate(def, now, bEnd, nanSamples)
	if resNaN.ValidCount != 2 {
		t.Errorf("Expected 2 valid samples, got %d", resNaN.ValidCount)
	}
	if resNaN.AvgValue == nil || *resNaN.AvgValue != 15.0 {
		t.Errorf("Expected AVG of 10 and 20 to be 15.0, got %v", resNaN.AvgValue)
	}

	// 23. Zero Valid Samples Handling
	zeroSamples := []model.RawData{
		{ProcessedValue: math.NaN(), ReceivedAt: now.Add(10 * time.Second), Quality: model.QualityBad},
	}
	resZero := eng.Calculate(def, now, bEnd, zeroSamples)
	if resZero.Value != nil {
		t.Errorf("Expected nil value when 0 valid samples, got %v", *resZero.Value)
	}
	if resZero.Quality != model.QualityBad {
		t.Errorf("Expected quality BAD when 0 valid samples, got %s", resZero.Quality)
	}
}
