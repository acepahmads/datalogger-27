package service_test

import (
	"math"
	"testing"

	"datalogger/internal/model"
	"datalogger/internal/service"
)

func TestISPUCalculator_Permen14OfficialWorkedExample(t *testing.T) {
	calc := service.NewISPUCalculator()

	// Page 11 worked example in Permen LHK No. 14 Tahun 2020:
	// Parameter: PM2.5, Measured concentration = 31.4 µg/m³
	// Interval: 15.5 - 55.4 µg/m³, Index: 50 - 100
	// Expected ISPU: ((100-50)/(55.4-15.5))*(31.4-15.5) + 50 = (50/39.9)*15.9 + 50 = 69.9248 -> rounded 70
	// Expected Category: Sedang
	out := calc.CalculateParameter(service.ParameterCalculationInput{
		PollutantCode:    model.PollutantPM25,
		ParameterID:      1,
		RawConcentration: 31.4,
		SourceUnit:       "µg/m³",
		ConversionFactor: 1.0,
		ValidSamples:     24,
		ExpectedSamples:  24,
	})

	if !out.IsEligible {
		t.Fatalf("Expected eligible result, got error: %s", out.ErrorMessage)
	}

	expectedFloat := ((100.0 - 50.0) / (55.4 - 15.5)) * (31.4 - 15.5) + 50.0
	if math.Abs(out.CalculatedISPU-expectedFloat) > 0.001 {
		t.Errorf("Expected CalculatedISPU %.4f, got %.4f", expectedFloat, out.CalculatedISPU)
	}

	if out.DisplayISPU != 70 {
		t.Errorf("Expected DisplayISPU 70, got %d", out.DisplayISPU)
	}

	if out.Category != model.CategorySedang {
		t.Errorf("Expected Category Sedang, got %s", out.Category)
	}
}

func TestISPUCalculator_AllSevenParametersBreakpoints(t *testing.T) {
	calc := service.NewISPUCalculator()

	testCases := []struct {
		Pollutant        model.ISPUParameter
		Concentration    float64
		ExpectedCategory model.ISPUCategory
		MinISPU          float64
		MaxISPU          float64
	}{
		// PM10
		{model.PollutantPM10, 25.0, model.CategoryBaik, 24.0, 26.0},
		{model.PollutantPM10, 100.0, model.CategorySedang, 74.0, 76.0},
		{model.PollutantPM10, 250.0, model.CategoryTidakSehat, 149.0, 151.0},
		{model.PollutantPM10, 385.0, model.CategorySangatTidakSehat, 249.0, 251.0},
		{model.PollutantPM10, 460.0, model.CategoryBerbahaya, 399.0, 401.0},

		// PM2.5
		{model.PollutantPM25, 7.75, model.CategoryBaik, 24.0, 26.0},
		{model.PollutantPM25, 35.45, model.CategorySedang, 74.0, 76.0},
		{model.PollutantPM25, 102.9, model.CategoryTidakSehat, 149.0, 151.0},
		{model.PollutantPM25, 200.4, model.CategorySangatTidakSehat, 249.0, 251.0},
		{model.PollutantPM25, 375.2, model.CategoryBerbahaya, 399.0, 401.0},

		// SO2
		{model.PollutantSO2, 26.0, model.CategoryBaik, 24.0, 26.0},
		{model.PollutantSO2, 116.0, model.CategorySedang, 74.0, 76.0},
		{model.PollutantSO2, 290.0, model.CategoryTidakSehat, 149.0, 151.0},
		{model.PollutantSO2, 600.0, model.CategorySangatTidakSehat, 249.0, 251.0},
		{model.PollutantSO2, 1000.0, model.CategoryBerbahaya, 399.0, 401.0},

		// CO
		{model.PollutantCO, 2000.0, model.CategoryBaik, 24.0, 26.0},
		{model.PollutantCO, 6000.0, model.CategorySedang, 74.0, 76.0},
		{model.PollutantCO, 11500.0, model.CategoryTidakSehat, 149.0, 151.0},
		{model.PollutantCO, 22500.0, model.CategorySangatTidakSehat, 249.0, 251.0},
		{model.PollutantCO, 37500.0, model.CategoryBerbahaya, 399.0, 401.0},

		// O3
		{model.PollutantO3, 60.0, model.CategoryBaik, 24.0, 26.0},
		{model.PollutantO3, 177.5, model.CategorySedang, 74.0, 76.0},
		{model.PollutantO3, 317.5, model.CategoryTidakSehat, 149.0, 151.0},
		{model.PollutantO3, 600.0, model.CategorySangatTidakSehat, 249.0, 251.0},
		{model.PollutantO3, 900.0, model.CategoryBerbahaya, 399.0, 401.0},

		// NO2
		{model.PollutantNO2, 40.0, model.CategoryBaik, 24.0, 26.0},
		{model.PollutantNO2, 140.0, model.CategorySedang, 74.0, 76.0},
		{model.PollutantNO2, 665.0, model.CategoryTidakSehat, 149.0, 151.0},
		{model.PollutantNO2, 1695.0, model.CategorySangatTidakSehat, 249.0, 251.0},
		{model.PollutantNO2, 2630.0, model.CategoryBerbahaya, 399.0, 401.0},

		// HC
		{model.PollutantHC, 22.5, model.CategoryBaik, 24.0, 26.0},
		{model.PollutantHC, 72.5, model.CategorySedang, 74.0, 76.0},
		{model.PollutantHC, 157.5, model.CategoryTidakSehat, 149.0, 151.0},
		{model.PollutantHC, 323.5, model.CategorySangatTidakSehat, 249.0, 251.0},
		{model.PollutantHC, 540.0, model.CategoryBerbahaya, 399.0, 401.0},
	}

	for _, tc := range testCases {
		out := calc.CalculateParameter(service.ParameterCalculationInput{
			PollutantCode:    tc.Pollutant,
			ParameterID:      1,
			RawConcentration: tc.Concentration,
			SourceUnit:       "µg/m³",
			ConversionFactor: 1.0,
			ValidSamples:     24,
			ExpectedSamples:  24,
		})

		if !out.IsEligible {
			t.Errorf("Pollutant %s (conc %.2f) calculation failed: %s", tc.Pollutant, tc.Concentration, out.ErrorMessage)
			continue
		}

		if out.Category != tc.ExpectedCategory {
			t.Errorf("Pollutant %s (conc %.2f) expected category %s, got %s (ISPU: %.2f)",
				tc.Pollutant, tc.Concentration, tc.ExpectedCategory, out.Category, out.CalculatedISPU)
		}

		if out.CalculatedISPU < tc.MinISPU || out.CalculatedISPU > tc.MaxISPU {
			t.Errorf("Pollutant %s (conc %.2f) ISPU %.2f out of expected range [%.1f, %.1f]",
				tc.Pollutant, tc.Concentration, out.CalculatedISPU, tc.MinISPU, tc.MaxISPU)
		}
	}
}

func TestISPUCalculator_ZeroAndBoundaryValues(t *testing.T) {
	calc := service.NewISPUCalculator()

	// Zero concentration: ISPU = 0, Baik
	outZero := calc.CalculateParameter(service.ParameterCalculationInput{
		PollutantCode:    model.PollutantPM25,
		RawConcentration: 0.0,
		SourceUnit:       "µg/m³",
		ValidSamples:     24,
		ExpectedSamples:  24,
	})
	if outZero.CalculatedISPU != 0 || outZero.Category != model.CategoryBaik {
		t.Errorf("Expected ISPU 0 Baik for zero conc, got %.2f %s", outZero.CalculatedISPU, outZero.Category)
	}

	// Exact Breakpoint boundary: PM2.5 = 15.5 µg/m³ -> ISPU 50 Baik
	outBp1 := calc.CalculateParameter(service.ParameterCalculationInput{
		PollutantCode:    model.PollutantPM25,
		RawConcentration: 15.5,
		SourceUnit:       "µg/m³",
		ValidSamples:     24,
		ExpectedSamples:  24,
	})
	if math.Abs(outBp1.CalculatedISPU-50.0) > 0.01 || outBp1.Category != model.CategoryBaik {
		t.Errorf("Expected boundary PM2.5 15.5 to yield ISPU 50 Baik, got %.2f %s", outBp1.CalculatedISPU, outBp1.Category)
	}

	// Exact Breakpoint boundary: PM2.5 = 55.4 µg/m³ -> ISPU 100 Sedang
	outBp2 := calc.CalculateParameter(service.ParameterCalculationInput{
		PollutantCode:    model.PollutantPM25,
		RawConcentration: 55.4,
		SourceUnit:       "µg/m³",
		ValidSamples:     24,
		ExpectedSamples:  24,
	})
	if math.Abs(outBp2.CalculatedISPU-100.0) > 0.01 || outBp2.Category != model.CategorySedang {
		t.Errorf("Expected boundary PM2.5 55.4 to yield ISPU 100 Sedang, got %.2f %s", outBp2.CalculatedISPU, outBp2.Category)
	}
}

func TestISPUCalculator_NegativeAndAboveRange(t *testing.T) {
	calc := service.NewISPUCalculator()

	// Negative concentration rejection
	outNeg := calc.CalculateParameter(service.ParameterCalculationInput{
		PollutantCode:    model.PollutantSO2,
		RawConcentration: -10.0,
		SourceUnit:       "µg/m³",
		ValidSamples:     24,
		ExpectedSamples:  24,
	})
	if outNeg.IsEligible {
		t.Errorf("Expected negative concentration to be rejected as ineligible, got eligible")
	}

	// Above range: PM2.5 = 600 µg/m³ (Above 500 max table value)
	outAbove := calc.CalculateParameter(service.ParameterCalculationInput{
		PollutantCode:    model.PollutantPM25,
		RawConcentration: 600.0,
		SourceUnit:       "µg/m³",
		ValidSamples:     24,
		ExpectedSamples:  24,
	})
	if !outAbove.AboveRange {
		t.Errorf("Expected AboveRange flag for PM2.5 conc 600")
	}
	if outAbove.Category != model.CategoryBerbahaya {
		t.Errorf("Expected Category Berbahaya for above range, got %s", outAbove.Category)
	}
	if outAbove.CalculatedISPU <= 500 {
		t.Errorf("Expected extrapolated ISPU > 500, got %.2f", outAbove.CalculatedISPU)
	}
}

func TestISPUCalculator_UnitConversions(t *testing.T) {
	calc := service.NewISPUCalculator()

	// mg/m³ to µg/m³: 0.05 mg/m³ = 50 µg/m³
	normVal, unit, err := calc.NormalizeConcentration(0.05, "mg/m³", "DIRECT", 0)
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}
	if unit != "µg/m³" || math.Abs(normVal-50.0) > 0.01 {
		t.Errorf("Expected 50 µg/m³, got %.2f %s", normVal, unit)
	}
}

func TestISPUCalculator_OverallAndDominantTieBreaker(t *testing.T) {
	calc := service.NewISPUCalculator()

	// Parameter 1: PM2.5 -> ISPU = 85 (Sedang)
	// Parameter 2: PM10  -> ISPU = 85 (Sedang) -> Tied with PM2.5!
	// Parameter 3: SO2   -> ISPU = 40 (Baik)
	results := []service.ParameterCalculationOutput{
		{
			PollutantCode:       model.PollutantPM10,
			CalculatedISPU:      85.0,
			Category:            model.CategorySedang,
			IsEligible:          true,
			CompletenessPercent: 100.0,
		},
		{
			PollutantCode:       model.PollutantPM25,
			CalculatedISPU:      85.0,
			Category:            model.CategorySedang,
			IsEligible:          true,
			CompletenessPercent: 100.0,
		},
		{
			PollutantCode:       model.PollutantSO2,
			CalculatedISPU:      40.0,
			Category:            model.CategoryBaik,
			IsEligible:          true,
			CompletenessPercent: 100.0,
		},
	}

	overall := calc.CalculateOverall(results, 75.0)

	if overall.OverallISPU != 85.0 {
		t.Errorf("Expected OverallISPU 85.0, got %.2f", overall.OverallISPU)
	}
	if overall.Category != model.CategorySedang {
		t.Errorf("Expected Category Sedang, got %s", overall.Category)
	}
	// Tie breaker: PM2.5 has higher priority than PM10
	if overall.DominantPollutant != model.PollutantPM25 {
		t.Errorf("Expected tie breaker to choose PM25, got %s", overall.DominantPollutant)
	}
	if len(overall.TiedPollutants) != 2 {
		t.Errorf("Expected 2 tied pollutants recorded, got %d", len(overall.TiedPollutants))
	}
	if overall.CalculationStatus != model.ISPUStatusValid {
		t.Errorf("Expected status VALID, got %s", overall.CalculationStatus)
	}
}

func TestISPUCalculator_DataCompletenessRegulatoryThresholds(t *testing.T) {
	calc := service.NewISPUCalculator()

	// Case 1: Completeness >= 75% -> VALID
	resultsValid := []service.ParameterCalculationOutput{
		{
			PollutantCode:       model.PollutantPM25,
			CalculatedISPU:      70.0,
			Category:            model.CategorySedang,
			IsEligible:          true,
			CompletenessPercent: 80.0,
		},
	}
	outValid := calc.CalculateOverall(resultsValid, 75.0)
	if outValid.CalculationStatus != model.ISPUStatusValid {
		t.Errorf("Expected VALID for 80%% completeness, got %s", outValid.CalculationStatus)
	}

	// Case 2: Completeness < 75% -> PROVISIONAL
	resultsProvisional := []service.ParameterCalculationOutput{
		{
			PollutantCode:       model.PollutantPM25,
			CalculatedISPU:      70.0,
			Category:            model.CategorySedang,
			IsEligible:          true,
			CompletenessPercent: 60.0,
		},
	}
	outProv := calc.CalculateOverall(resultsProvisional, 75.0)
	if outProv.CalculationStatus != model.ISPUStatusProvisional {
		t.Errorf("Expected PROVISIONAL for 60%% completeness, got %s", outProv.CalculationStatus)
	}

	// Case 3: Empty results -> INVALID
	outEmpty := calc.CalculateOverall(nil, 75.0)
	if outEmpty.CalculationStatus != model.ISPUStatusInvalid {
		t.Errorf("Expected INVALID for nil results, got %s", outEmpty.CalculationStatus)
	}
}
