package calibration_test

import (
	"math"
	"testing"

	"datalogger/internal/calibration"
)

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestRegression_1PointOffsetAndSpan(t *testing.T) {
	// 1-Point Offset: pH 7.00 buffer reads 6.82 -> y = x + 0.18
	resOffset, err := calibration.CalculateRegression(
		calibration.ModelOffset1Pt,
		[]calibration.CalibrationPoint{
			{Index: 1, Active: true, Label: "Buffer pH 7.00", RawX: 6.82, RefY: 7.00},
		},
		0.995,
	)
	if err != nil {
		t.Fatalf("unexpected error for 1-point offset: %v", err)
	}
	if !almostEqual(resOffset.C1, 1.0, 1e-6) || !almostEqual(resOffset.C0, 0.18, 1e-6) {
		t.Fatalf("expected c1=1.0, c0=0.18, got c1=%f, c0=%f", resOffset.C1, resOffset.C0)
	}
	if resOffset.R2 != nil {
		t.Fatalf("expected R2=nil for 1-point calibration, got %v", *resOffset.R2)
	}
	if !resOffset.PassedR2 {
		t.Fatalf("expected PassedR2=true for 1-point calibration")
	}

	// 1-Point Span: DO air-saturated 8.26 mg/L reads 7.90 -> y = 1.04557*x
	resSpan, err := calibration.CalculateRegression(
		calibration.ModelSpan1Pt,
		[]calibration.CalibrationPoint{
			{Index: 1, Active: true, Label: "DO Jenuh Udara", RawX: 7.90, RefY: 8.26},
		},
		0.995,
	)
	if err != nil {
		t.Fatalf("unexpected error for 1-point span: %v", err)
	}
	expectedSlope := 8.26 / 7.90
	if !almostEqual(resSpan.C1, expectedSlope, 1e-5) || !almostEqual(resSpan.C0, 0.0, 1e-6) {
		t.Fatalf("expected slope=%f, c0=0, got c1=%f, c0=%f", expectedSlope, resSpan.C1, resSpan.C0)
	}
}

func TestRegression_2To10PointLinearAndPerParamMinR2(t *testing.T) {
	// 5-Point AQMS SO2 Linear calibration
	pts5 := []calibration.CalibrationPoint{
		{Index: 1, Active: true, RawX: 0.4, RefY: 0.0},
		{Index: 2, Active: true, RawX: 98.2, RefY: 100.0},
		{Index: 3, Active: true, RawX: 245.5, RefY: 250.0},
		{Index: 4, Active: true, RawX: 490.8, RefY: 500.0},
		{Index: 5, Active: true, RawX: 785.0, RefY: 800.0},
	}

	res5, err := calibration.CalculateRegression(calibration.ModelLinear, pts5, 0.998)
	if err != nil {
		t.Fatalf("unexpected error for 5-point linear: %v", err)
	}
	if res5.R2 == nil || *res5.R2 < 0.9999 {
		t.Fatalf("expected high R2 >= 0.9999, got %v", res5.R2)
	}
	if !res5.PassedR2 || !almostEqual(res5.MinR2, 0.998, 1e-6) {
		t.Fatalf("expected PassedR2=true with MinR2=0.998")
	}

	// 10-Point Linear calibration across full range
	pts10 := make([]calibration.CalibrationPoint, 10)
	for i := 0; i < 10; i++ {
		ref := float64(i * 50) // 0 .. 450
		raw := (ref - 1.5) / 1.02
		pts10[i] = calibration.CalibrationPoint{
			Index:  i + 1,
			Active: true,
			RawX:   raw,
			RefY:   ref,
		}
	}
	res10, err := calibration.CalculateRegression(calibration.ModelLinear, pts10, 0.999)
	if err != nil {
		t.Fatalf("unexpected error for 10-point linear: %v", err)
	}
	if res10.ActiveCount != 10 || !almostEqual(res10.C1, 1.02, 1e-5) || !almostEqual(res10.C0, 1.5, 1e-5) {
		t.Fatalf("unexpected 10-point linear fit: count=%d c1=%f c0=%f", res10.ActiveCount, res10.C1, res10.C0)
	}

	// Exceeding 10 points must be rejected
	pts11 := append(pts10, calibration.CalibrationPoint{Index: 11, Active: true, RawX: 500, RefY: 511.5})
	if _, err := calibration.CalculateRegression(calibration.ModelLinear, pts11, 0.995); err == nil {
		t.Fatalf("expected error when > 10 points are supplied")
	}

	// Per-parameter MinR2 threshold check: noisy curve with R2 ~ 0.991 passes when MinR2=0.985, fails when MinR2=0.995
	noisyPts := []calibration.CalibrationPoint{
		{Index: 1, Active: true, RawX: 0, RefY: 0},
		{Index: 2, Active: true, RawX: 25, RefY: 31},
		{Index: 3, Active: true, RawX: 50, RefY: 46},
		{Index: 4, Active: true, RawX: 75, RefY: 79},
		{Index: 5, Active: true, RawX: 100, RefY: 99},
	}
	resStrict, err := calibration.CalculateRegression(calibration.ModelLinear, noisyPts, 0.995)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resStrict.PassedR2 {
		t.Fatalf("expected noisy curve (R2=%f) to fail strict MinR2=0.995", *resStrict.R2)
	}

	resRelaxed, err := calibration.CalculateRegression(calibration.ModelLinear, noisyPts, 0.980)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resRelaxed.PassedR2 {
		t.Fatalf("expected noisy curve (R2=%f) to pass per-parameter MinR2=0.980", *resRelaxed.R2)
	}
}

func TestRegression_QuadraticAndOutlierExclusion(t *testing.T) {
	// Non-linear optical turbidity / COD curve: y = 0.0012*x^2 + 0.94*x + 0.50
	// Include 1 unchecked outlier at index 4
	trueCurve := func(x float64) float64 {
		return 0.0012*x*x + 0.94*x + 0.50
	}
	pts := []calibration.CalibrationPoint{
		{Index: 1, Active: true, RawX: 0, RefY: trueCurve(0)},
		{Index: 2, Active: true, RawX: 50, RefY: trueCurve(50)},
		{Index: 3, Active: true, RawX: 100, RefY: trueCurve(100)},
		{Index: 4, Active: false, RawX: 150, RefY: 999.0}, // unchecked outlier
		{Index: 5, Active: true, RawX: 200, RefY: trueCurve(200)},
		{Index: 6, Active: true, RawX: 300, RefY: trueCurve(300)},
		{Index: 7, Active: true, RawX: 400, RefY: trueCurve(400)},
	}

	res, err := calibration.CalculateRegression(calibration.ModelQuadratic, pts, 0.999)
	if err != nil {
		t.Fatalf("unexpected error for quadratic fit: %v", err)
	}
	if res.ActiveCount != 6 || res.TotalCount != 7 {
		t.Fatalf("expected ActiveCount=6, TotalCount=7, got %d/%d", res.ActiveCount, res.TotalCount)
	}
	if !almostEqual(res.C2, 0.0012, 1e-6) || !almostEqual(res.C1, 0.94, 1e-5) || !almostEqual(res.C0, 0.50, 1e-4) {
		t.Fatalf("quadratic coefficients mismatch: c2=%f c1=%f c0=%f", res.C2, res.C1, res.C0)
	}
	if res.R2 == nil || *res.R2 < 0.999999 || !res.PassedR2 {
		t.Fatalf("expected near-1.0 R2 when outlier is unchecked, got %v", res.R2)
	}
	// Verify inactive point still gets predicted Y & residual computed for diagnostic table
	expectedOutlierPred := trueCurve(150)
	if !almostEqual(res.Points[3].PredY, expectedOutlierPred, 1e-4) {
		t.Fatalf("expected inactive point PredY=%f, got %f", expectedOutlierPred, res.Points[3].PredY)
	}
}

