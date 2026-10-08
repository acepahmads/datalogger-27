package communication

import (
	"errors"
	"math"
	"testing"
	"time"

	"datalogger/internal/model"
)

func TestAnomalyHoldTracker_NormalReadings(t *testing.T) {
	tracker := NewAnomalyHoldTracker()

	param := &model.Parameter{
		ID:                    10,
		ParameterCode:         "TEMP_01",
		HoldLastValueEnabled: true,
		HoldLastValueSeconds: 120,
		Formula:               "x * 1.8 + 32", // Celsius to Fahrenheit
	}

	// 1. Initial normal reading: 25.0 °C
	effVal, formulaVal, isHeld, quality, err := tracker.ProcessReading(param, 250, 25.0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isHeld {
		t.Errorf("expected isHeld=false on normal reading")
	}
	if quality != model.QualityGood {
		t.Errorf("expected QualityGood, got %v", quality)
	}
	if effVal != 25.0 {
		t.Errorf("expected effVal=25.0, got %v", effVal)
	}
	if formulaVal == nil || math.Abs(*formulaVal-77.0) > 0.001 {
		t.Errorf("expected formulaVal=77.0, got %v", formulaVal)
	}
}

func TestAnomalyHoldTracker_HoldLastValueOnTimeout(t *testing.T) {
	tracker := NewAnomalyHoldTracker()

	param := &model.Parameter{
		ID:                    20,
		ParameterCode:         "PRESSURE_01",
		HoldLastValueEnabled: true,
		HoldLastValueSeconds: 2, // 2 seconds for test speed
		Formula:               "x * 10",
	}

	// 1. Seed good reading: 5.0 bar
	_, _, _, _, err := tracker.ProcessReading(param, 50, 5.0, nil)
	if err != nil {
		t.Fatalf("unexpected seed error: %v", err)
	}

	// 2. Sensor timeout / read error occurs!
	timeoutErr := errors.New("i/o timeout")
	effVal, formulaVal, isHeld, quality, err := tracker.ProcessReading(param, 0, 0, timeoutErr)
	if err != nil {
		t.Fatalf("expected error to be suppressed during hold grace period, got %v", err)
	}
	if !isHeld {
		t.Errorf("expected isHeld=true during timeout within 2s window")
	}
	if effVal != 5.0 {
		t.Errorf("expected held value=5.0, got %v", effVal)
	}
	if formulaVal == nil || *formulaVal != 50.0 {
		t.Errorf("expected formula value on held value=50.0, got %v", formulaVal)
	}
	if quality != model.QualityUncertain {
		t.Errorf("expected QualityUncertain for held value, got %v", quality)
	}

	// 3. Sensor recovers before grace period expires!
	effVal, formulaVal, isHeld, quality, err = tracker.ProcessReading(param, 52, 5.2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isHeld {
		t.Errorf("expected isHeld=false after recovery")
	}
	if effVal != 5.2 {
		t.Errorf("expected recovered value 5.2, got %v", effVal)
	}
	if formulaVal == nil || math.Abs(*formulaVal-52.0) > 0.001 {
		t.Errorf("expected formula 52.0, got %v", formulaVal)
	}
}

func TestAnomalyHoldTracker_HoldLastValueOnOutOfBoundsAneh(t *testing.T) {
	tracker := NewAnomalyHoldTracker()

	minV := 0.0
	maxV := 100.0
	param := &model.Parameter{
		ID:                    30,
		ParameterCode:         "HUMIDITY",
		MinValue:              &minV,
		MaxValue:              &maxV,
		HoldLastValueEnabled: true,
		HoldLastValueSeconds: 120,
	}

	// 1. Seed valid reading: 65.0 %RH
	_, _, _, _, _ = tracker.ProcessReading(param, 650, 65.0, nil)

	// 2. Sensor sends "aneh" / glitch spike value: 9999.0 (exceeds max 100)
	effVal, _, isHeld, quality, err := tracker.ProcessReading(param, 99990, 9999.0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isHeld {
		t.Errorf("expected isHeld=true for spike exceeding MaxValue")
	}
	if effVal != 65.0 {
		t.Errorf("expected held last good value 65.0, got %v", effVal)
	}
	if quality != model.QualityUncertain {
		t.Errorf("expected QualityUncertain, got %v", quality)
	}
}

func TestAnomalyHoldTracker_GracePeriodExpiry(t *testing.T) {
	tracker := NewAnomalyHoldTracker()

	param := &model.Parameter{
		ID:                    40,
		ParameterCode:         "FLOW_RATE",
		HoldLastValueEnabled: true,
		HoldLastValueSeconds: 1, // 1 second grace window
	}

	// 1. Seed valid reading
	_, _, _, _, _ = tracker.ProcessReading(param, 100, 10.0, nil)

	// 2. First failure: should be held
	effVal, _, isHeld, _, _ := tracker.ProcessReading(param, 0, 0, errors.New("sensor disconnected"))
	if !isHeld || effVal != 10.0 {
		t.Fatalf("expected held last good value 10.0, got %v (isHeld=%v)", effVal, isHeld)
	}

	// 3. Wait for grace period to expire (> 1 second)
	time.Sleep(1100 * time.Millisecond)

	// 4. Next failure after > 1 second: MUST COMMIT REAL SENSOR STATE / ERROR!
	_, _, isHeld, quality, err := tracker.ProcessReading(param, 0, 0, errors.New("sensor disconnected"))
	if isHeld {
		t.Errorf("expected isHeld=false after grace period expired")
	}
	if err == nil {
		t.Errorf("expected real error to be returned after grace period expired")
	}
	if quality != model.QualityBad {
		t.Errorf("expected QualityBad, got %v", quality)
	}
}

func TestAnomalyHoldTracker_DisabledFeature(t *testing.T) {
	tracker := NewAnomalyHoldTracker()

	param := &model.Parameter{
		ID:                    50,
		ParameterCode:         "VOLT_L2",
		HoldLastValueEnabled: false, // Disabled!
	}

	// 1. Seed valid reading
	_, _, _, _, _ = tracker.ProcessReading(param, 220, 220.0, nil)

	// 2. Read error occurs: should immediately return error without holding
	_, _, isHeld, quality, err := tracker.ProcessReading(param, 0, 0, errors.New("hardware fault"))
	if isHeld {
		t.Errorf("expected isHeld=false when feature is disabled")
	}
	if err == nil {
		t.Errorf("expected error to pass through immediately")
	}
	if quality != model.QualityBad {
		t.Errorf("expected QualityBad, got %v", quality)
	}
}
