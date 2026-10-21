package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"datalogger/internal/model"
)

// ISPUBreakpoint represents a single regulatory concentration-to-index range.
type ISPUBreakpoint struct {
	Ib float64 // Lower ISPU boundary
	Ia float64 // Upper ISPU boundary
	Xb float64 // Lower ambient concentration boundary (µg/m³)
	Xa float64 // Upper ambient concentration boundary (µg/m³)
}

// OfficialBreakpoints stores the official 24-hour conversion breakpoints from Lampiran I Tabel 1
// of Peraturan Menteri LHK No. 14 Tahun 2020.
var OfficialBreakpoints = map[model.ISPUParameter][]ISPUBreakpoint{
	model.PollutantPM10: {
		{Ib: 0, Ia: 50, Xb: 0, Xa: 50},
		{Ib: 50, Ia: 100, Xb: 50, Xa: 150},
		{Ib: 100, Ia: 200, Xb: 150, Xa: 350},
		{Ib: 200, Ia: 300, Xb: 350, Xa: 420},
		{Ib: 300, Ia: 500, Xb: 420, Xa: 500},
	},
	model.PollutantPM25: {
		{Ib: 0, Ia: 50, Xb: 0, Xa: 15.5},
		{Ib: 50, Ia: 100, Xb: 15.5, Xa: 55.4},
		{Ib: 100, Ia: 200, Xb: 55.4, Xa: 150.4},
		{Ib: 200, Ia: 300, Xb: 150.4, Xa: 250.4},
		{Ib: 300, Ia: 500, Xb: 250.4, Xa: 500},
	},
	model.PollutantSO2: {
		{Ib: 0, Ia: 50, Xb: 0, Xa: 52},
		{Ib: 50, Ia: 100, Xb: 52, Xa: 180},
		{Ib: 100, Ia: 200, Xb: 180, Xa: 400},
		{Ib: 200, Ia: 300, Xb: 400, Xa: 800},
		{Ib: 300, Ia: 500, Xb: 800, Xa: 1200},
	},
	model.PollutantCO: {
		{Ib: 0, Ia: 50, Xb: 0, Xa: 4000},
		{Ib: 50, Ia: 100, Xb: 4000, Xa: 8000},
		{Ib: 100, Ia: 200, Xb: 8000, Xa: 15000},
		{Ib: 200, Ia: 300, Xb: 15000, Xa: 30000},
		{Ib: 300, Ia: 500, Xb: 30000, Xa: 45000},
	},
	model.PollutantO3: {
		{Ib: 0, Ia: 50, Xb: 0, Xa: 120},
		{Ib: 50, Ia: 100, Xb: 120, Xa: 235},
		{Ib: 100, Ia: 200, Xb: 235, Xa: 400},
		{Ib: 200, Ia: 300, Xb: 400, Xa: 800},
		{Ib: 300, Ia: 500, Xb: 800, Xa: 1000},
	},
	model.PollutantNO2: {
		{Ib: 0, Ia: 50, Xb: 0, Xa: 80},
		{Ib: 50, Ia: 100, Xb: 80, Xa: 200},
		{Ib: 100, Ia: 200, Xb: 200, Xa: 1130},
		{Ib: 200, Ia: 300, Xb: 1130, Xa: 2260},
		{Ib: 300, Ia: 500, Xb: 2260, Xa: 3000},
	},
	model.PollutantHC: {
		{Ib: 0, Ia: 50, Xb: 0, Xa: 45},
		{Ib: 50, Ia: 100, Xb: 45, Xa: 100},
		{Ib: 100, Ia: 200, Xb: 100, Xa: 215},
		{Ib: 200, Ia: 300, Xb: 215, Xa: 432},
		{Ib: 300, Ia: 500, Xb: 432, Xa: 648},
	},
}

// ParameterCalculationInput holds the raw input for calculating a single pollutant ISPU.
type ParameterCalculationInput struct {
	PollutantCode    model.ISPUParameter
	ParameterID      uint
	RawConcentration float64
	SourceUnit       string
	ConversionFactor float64
	ConversionRule   string
	ValidSamples     int
	ExpectedSamples  int
	MinCompleteness  float64
}

// ParameterCalculationOutput holds the detailed mathematical calculation result.
type ParameterCalculationOutput struct {
	PollutantCode           model.ISPUParameter `json:"pollutant_code"`
	ParameterID             uint                `json:"parameter_id"`
	RawConcentration        float64             `json:"raw_concentration"`
	SourceUnit              string              `json:"source_unit"`
	NormalizedConcentration float64             `json:"normalized_concentration"`
	NormalizedUnit          string              `json:"normalized_unit"`
	CalculatedISPU          float64             `json:"calculated_ispu"`
	DisplayISPU             int                 `json:"display_ispu"`
	Category                model.ISPUCategory  `json:"category"`
	Ia                      float64             `json:"ia"`
	Ib                      float64             `json:"ib"`
	Xa                      float64             `json:"xa"`
	Xb                      float64             `json:"xb"`
	CompletenessPercent     float64             `json:"completeness_percent"`
	ValidSamples            int                 `json:"valid_samples"`
	ExpectedSamples         int                 `json:"expected_samples"`
	Formula                  string              `json:"formula"`
	IsEligible               bool                `json:"is_eligible"`
	AboveRange               bool                `json:"above_range"`
	AggregationConcentration  float64             `json:"aggregation_concentration"`
	AggregationISPU           float64             `json:"aggregation_ispu"`
	AggregationCategory       model.ISPUCategory  `json:"aggregation_category"`
	SourceAggregationCount    int                 `json:"source_aggregation_count"`
	SourceLatestIdentifier    string              `json:"source_latest_identifier"`
	SourceAggregationInterval int                 `json:"source_aggregation_interval"`
	SourceAggregationFunction string              `json:"source_aggregation_function"`
	PeakConcentration         float64             `json:"peak_concentration"`
	PeakTimestamp             *time.Time          `json:"peak_timestamp,omitempty"`
	PeakBucketIdentifier      string              `json:"peak_bucket_identifier,omitempty"`
	MinConcentration          float64             `json:"min_concentration"`
	DetailsJSON               string              `json:"details_json,omitempty"`
	ErrorMessage              string              `json:"error_message,omitempty"`
}

// OverallISPUOutput holds the synthesized station result and dominant pollutant.
type OverallISPUOutput struct {
	OverallISPU         float64                      `json:"overall_ispu"`
	DisplayISPU         int                          `json:"display_ispu"`
	Category            model.ISPUCategory           `json:"category"`
	DominantPollutant   model.ISPUParameter          `json:"dominant_pollutant"`
	TiedPollutants      []model.ISPUParameter        `json:"tied_pollutants,omitempty"`
	CalculationStatus   model.ISPUStatus             `json:"calculation_status"`
	CompletenessPercent float64                      `json:"completeness_percent"`
	AggregationISPU     float64                      `json:"aggregation_ispu"`
	AggregationCategory model.ISPUCategory           `json:"aggregation_category"`
	AggregationDominant model.ISPUParameter          `json:"aggregation_dominant,omitempty"`
	Parameters          []ParameterCalculationOutput `json:"parameters"`
	Notes               string                       `json:"notes,omitempty"`
}

// ISPUCalculator provides mathematical evaluation strictly conforming to Permen LHK No. 14 Tahun 2020.
type ISPUCalculator struct{}

// NewISPUCalculator creates a new calculator instance.
func NewISPUCalculator() *ISPUCalculator {
	return &ISPUCalculator{}
}

// NormalizeConcentration converts source concentration to µg/m³ based on unit conversion rules.
func (c *ISPUCalculator) NormalizeConcentration(val float64, unit, rule string, customFactor float64) (float64, string, error) {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0, "", fmt.Errorf("invalid concentration value: NaN/Inf")
	}
	if val < 0 {
		return 0, "", fmt.Errorf("negative concentration rejected: %.2f", val)
	}

	cleanUnit := strings.TrimSpace(unit)
	u := strings.ToLower(cleanUnit)

	// Direct or µg/m³
	if u == "µg/m³" || u == "ug/m3" || u == "ug/m^3" || u == "µg/nm³" || u == "ug/nm3" {
		factor := 1.0
		if customFactor > 0 && customFactor != 1.0 {
			factor = customFactor
		}
		return val * factor, "µg/m³", nil
	}

	// mg/m³ -> µg/m³ (1 mg/m³ = 1000 µg/m³)
	if u == "mg/m³" || u == "mg/m3" || u == "mg/m^3" {
		factor := 1000.0
		if customFactor > 0 {
			factor = customFactor
		}
		return val * factor, "µg/m³", nil
	}

	// ppm or ppb with custom factor or rule
	if customFactor > 0 {
		return val * customFactor, "µg/m³", nil
	}

	// Default fallback: direct
	return val, "µg/m³", nil
}

// CalculateParameter evaluates Lampiran I Persamaan 1 for a single pollutant:
// I = ((Ia - Ib) / (Xa - Xb)) * (Xx - Xb) + Ib
func (c *ISPUCalculator) CalculateParameter(input ParameterCalculationInput) ParameterCalculationOutput {
	out := ParameterCalculationOutput{
		PollutantCode:    input.PollutantCode,
		ParameterID:      input.ParameterID,
		RawConcentration: input.RawConcentration,
		SourceUnit:       input.SourceUnit,
		NormalizedUnit:   "µg/m³",
		ExpectedSamples:  input.ExpectedSamples,
		ValidSamples:     input.ValidSamples,
		Category:         model.CategoryUnavailable,
	}

	// Calculate temporal completeness
	if input.ExpectedSamples > 0 {
		out.CompletenessPercent = (float64(input.ValidSamples) / float64(input.ExpectedSamples)) * 100.0
	} else if input.ValidSamples > 0 {
		out.CompletenessPercent = 100.0
	}

	// Negative value check
	if input.RawConcentration < 0 {
		out.ErrorMessage = fmt.Sprintf("invalid negative concentration: %.2f", input.RawConcentration)
		return out
	}

	// Unit normalization
	normVal, normUnit, err := c.NormalizeConcentration(input.RawConcentration, input.SourceUnit, input.ConversionRule, input.ConversionFactor)
	if err != nil {
		out.ErrorMessage = err.Error()
		return out
	}
	out.NormalizedConcentration = normVal
	out.NormalizedUnit = normUnit

	// Lookup breakpoints for pollutant
	breakpoints, exists := OfficialBreakpoints[input.PollutantCode]
	if !exists {
		out.ErrorMessage = fmt.Sprintf("unknown or unregulated ISPU pollutant: %s", input.PollutantCode)
		return out
	}

	// Missing / unavailable data check: expected samples > 0 but valid samples == 0
	if input.ExpectedSamples > 0 && input.ValidSamples == 0 {
		out.Category = model.CategoryUnavailable
		out.CalculatedISPU = 0
		out.DisplayISPU = 0
		out.IsEligible = false
		out.ErrorMessage = "no valid observations in measurement window"
		return out
	}

	// Handle measured zero (Xx = 0 with ValidSamples > 0)
	if normVal == 0 {
		bp := breakpoints[0]
		out.CalculatedISPU = 0
		out.DisplayISPU = 0
		out.Category = model.CategoryBaik
		out.Ia = bp.Ia
		out.Ib = bp.Ib
		out.Xa = bp.Xa
		out.Xb = bp.Xb
		out.Formula = fmt.Sprintf("((%.1f - %.1f) / (%.1f - %.1f)) * (0 - %.1f) + %.1f = 0", bp.Ia, bp.Ib, bp.Xa, bp.Xb, bp.Xb, bp.Ib)
		out.IsEligible = true
		return out
	}

	// Find the matching breakpoint interval
	var matched *ISPUBreakpoint
	for idx := range breakpoints {
		bp := &breakpoints[idx]
		if idx == 0 {
			if normVal >= bp.Xb && normVal <= bp.Xa {
				matched = bp
				break
			}
		} else {
			if normVal > bp.Xb && normVal <= bp.Xa {
				matched = bp
				break
			}
		}
	}

	// If concentration exceeds highest breakpoint (Above Table Maximum)
	if matched == nil {
		lastBp := breakpoints[len(breakpoints)-1]
		if normVal > lastBp.Xa {
			matched = &lastBp
			out.AboveRange = true
		} else {
			out.ErrorMessage = fmt.Sprintf("concentration %.2f out of breakpoint bounds", normVal)
			return out
		}
	}

	// Apply regulatory equation:
	// I = ((Ia - Ib) / (Xa - Xb)) * (Xx - Xb) + Ib
	ia := matched.Ia
	ib := matched.Ib
	xa := matched.Xa
	xb := matched.Xb

	ispu := ((ia - ib) / (xa - xb)) * (normVal - xb) + ib

	// Lampiran II Category determination based on ISPU value
	cat := GetCategoryFromISPU(ispu)

	out.CalculatedISPU = ispu
	out.DisplayISPU = int(math.Round(ispu))
	out.Category = cat
	out.Ia = ia
	out.Ib = ib
	out.Xa = xa
	out.Xb = xb
	out.Formula = fmt.Sprintf("((%.1f - %.1f) / (%.1f - %.1f)) * (%.2f - %.1f) + %.1f = %.2f (display: %d)",
		ia, ib, xa, xb, normVal, xb, ib, ispu, out.DisplayISPU)
	out.IsEligible = true

	return out
}

// CalculateOverall evaluates station-wide ISPU, dominant pollutant, and validity.
func (c *ISPUCalculator) CalculateOverall(results []ParameterCalculationOutput, minCompleteness float64) OverallISPUOutput {
	if minCompleteness <= 0 {
		minCompleteness = 75.0 // Permen LHK 14/2020 Pasal 7 ayat (3) default
	}

	out := OverallISPUOutput{
		Category:          model.CategoryUnavailable,
		CalculationStatus: model.ISPUStatusInvalid,
		Parameters:        results,
	}

	if len(results) == 0 {
		out.Notes = "no parameter results available"
		return out
	}

	var eligibleResults []ParameterCalculationOutput
	var sumCompleteness float64
	var maxISPU float64 = -1
	var dominantPollutants []model.ISPUParameter

	for _, pr := range results {
		sumCompleteness += pr.CompletenessPercent
		if !pr.IsEligible || pr.ErrorMessage != "" {
			continue
		}
		eligibleResults = append(eligibleResults, pr)

		if pr.CalculatedISPU > maxISPU {
			maxISPU = pr.CalculatedISPU
			dominantPollutants = []model.ISPUParameter{pr.PollutantCode}
		} else if pr.CalculatedISPU == maxISPU {
			dominantPollutants = append(dominantPollutants, pr.PollutantCode)
		}
	}

	avgCompleteness := sumCompleteness / float64(len(results))
	out.CompletenessPercent = avgCompleteness

	if len(eligibleResults) == 0 {
		out.Notes = "no eligible parameters passed quality/bounds check"
		out.CalculationStatus = model.ISPUStatusInvalid
		return out
	}

	out.OverallISPU = maxISPU
	out.DisplayISPU = int(math.Round(maxISPU))
	out.Category = GetCategoryFromISPU(maxISPU)

	// Dominant pollutant tie-breaking hierarchy: PM2.5 > PM10 > O3 > NO2 > SO2 > CO > HC
	if len(dominantPollutants) == 1 {
		out.DominantPollutant = dominantPollutants[0]
	} else if len(dominantPollutants) > 1 {
		out.TiedPollutants = dominantPollutants
		out.DominantPollutant = resolveDominantTie(dominantPollutants)
	}

	// Regulatory Validity determination:
	// Pasal 7 ayat (3): Completeness >= 75% for VALID
	if avgCompleteness >= minCompleteness {
		out.CalculationStatus = model.ISPUStatusValid
	} else if avgCompleteness > 0 {
		out.CalculationStatus = model.ISPUStatusProvisional
		out.Notes = fmt.Sprintf("provisional result: data completeness (%.1f%%) is below regulatory threshold (%.1f%%)",
			avgCompleteness, minCompleteness)
	} else {
		out.CalculationStatus = model.ISPUStatusInvalid
		out.Notes = "insufficient telemetry coverage for regulatory evaluation"
	}

	return out
}

// GetCategoryFromISPU returns the official category from Lampiran II based on rounded integer ISPU.
func GetCategoryFromISPU(ispu float64) model.ISPUCategory {
	rounded := math.Round(ispu)
	switch {
	case rounded <= 50:
		return model.CategoryBaik
	case rounded <= 100:
		return model.CategorySedang
	case rounded <= 200:
		return model.CategoryTidakSehat
	case rounded <= 300:
		return model.CategorySangatTidakSehat
	default:
		return model.CategoryBerbahaya
	}
}

// GetCategoryColor returns standard Hex color and SCADA class according to Lampiran II.
func GetCategoryColor(category model.ISPUCategory) (string, string) {
	switch category {
	case model.CategoryBaik:
		return "#10B981", "ispu-baik" // Green / Hijau
	case model.CategorySedang:
		return "#3B82F6", "ispu-sedang" // Blue / Biru
	case model.CategoryTidakSehat:
		return "#EAB308", "ispu-tidak-sehat" // Yellow / Kuning
	case model.CategorySangatTidakSehat:
		return "#EF4444", "ispu-sangat-tidak-sehat" // Red / Merah
	case model.CategoryBerbahaya:
		return "#1E1B4B", "ispu-berbahaya" // Dark Navy / Hitam
	default:
		return "#64748B", "ispu-unavailable" // Slate / Gray
	}
}

// resolveDominantTie resolves ties by official hierarchy order: PM2.5 > PM10 > O3 > NO2 > SO2 > CO > HC
func resolveDominantTie(tied []model.ISPUParameter) model.ISPUParameter {
	order := []model.ISPUParameter{
		model.PollutantPM25,
		model.PollutantPM10,
		model.PollutantO3,
		model.PollutantNO2,
		model.PollutantSO2,
		model.PollutantCO,
		model.PollutantHC,
	}

	for _, candidate := range order {
		for _, p := range tied {
			if p == candidate {
				return p
			}
		}
	}

	return tied[0]
}
