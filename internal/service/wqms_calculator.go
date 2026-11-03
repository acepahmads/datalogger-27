package service

import (
	"fmt"
	"math"
	"strings"

	"datalogger/internal/model"
)

// WQMSStandardRule defines the mathematical evaluation rule for a parameter under Kepmen LH No. 115 Tahun 2003.
const (
	WQMSRuleMax       = "MAX"        // Standard upper limit (Ci / Lij)
	WQMSRuleRange     = "RANGE"      // Bounded range [Lmin, Lmax], e.g. pH 6.0 - 9.0
	WQMSRuleInverseDO = "INVERSE_DO" // Minimum required DO evaluated via deficit from Cim saturation
)

// WQMSClassStandard represents the regulatory threshold of a parameter for a given Water Class (PP No. 22 Tahun 2021 Lampiran VI).
type WQMSClassStandard struct {
	ParamCode   model.WQMSParamCode `json:"param_code"`
	DisplayName string              `json:"display_name"`
	ShortLabel  string              `json:"short_label"`
	Unit        string              `json:"unit"`
	Rule        string              `json:"rule"` // MAX, RANGE, INVERSE_DO
	MinLimit    *float64            `json:"min_limit,omitempty"`
	MaxLimit    *float64            `json:"max_limit,omitempty"`
	Label       string              `json:"label"`
}

func ptrFloat(v float64) *float64 {
	return &v
}

// OfficialWQMSStandards stores Lampiran VI PP No. 22 Tahun 2021 surface/river water standards across Classes I-IV.
var OfficialWQMSStandards = map[model.WQMSWaterClass]map[model.WQMSParamCode]WQMSClassStandard{
	model.WaterClass1: {
		model.WQMSParamPH:        {ParamCode: model.WQMSParamPH, DisplayName: "pH (Derajat Keasaman)", ShortLabel: "pH", Unit: "pH", Rule: WQMSRuleRange, MinLimit: ptrFloat(6.0), MaxLimit: ptrFloat(9.0), Label: "6.0 – 9.0"},
		model.WQMSParamDO:        {ParamCode: model.WQMSParamDO, DisplayName: "DO (Oksigen Terlarut)", ShortLabel: "DO", Unit: "mg/L", Rule: WQMSRuleInverseDO, MinLimit: ptrFloat(6.0), Label: "≥ 6.0 mg/L"},
		model.WQMSParamBOD:       {ParamCode: model.WQMSParamBOD, DisplayName: "BOD5", ShortLabel: "BOD", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(2.0), Label: "2.0 mg/L"},
		model.WQMSParamCOD:       {ParamCode: model.WQMSParamCOD, DisplayName: "COD", ShortLabel: "COD", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(10.0), Label: "10 mg/L"},
		model.WQMSParamTSS:       {ParamCode: model.WQMSParamTSS, DisplayName: "TSS (Padatan Tersuspensi)", ShortLabel: "TSS", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(40.0), Label: "40 mg/L"},
		model.WQMSParamNO3:       {ParamCode: model.WQMSParamNO3, DisplayName: "Nitrat (NO3-N)", ShortLabel: "NO3-N", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(10.0), Label: "10 mg/L"},
		model.WQMSParamNH3:       {ParamCode: model.WQMSParamNH3, DisplayName: "Amonia Total (NH3-N)", ShortLabel: "NH3-N", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(0.1), Label: "0.1 mg/L"},
		model.WQMSParamTDS:       {ParamCode: model.WQMSParamTDS, DisplayName: "TDS (Padatan Terlarut)", ShortLabel: "TDS", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(1000.0), Label: "1000 mg/L"},
		model.WQMSParamPO4:       {ParamCode: model.WQMSParamPO4, DisplayName: "Total Fosfat (PO4-P)", ShortLabel: "PO4-P", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(0.2), Label: "0.2 mg/L"},
		model.WQMSParamTurbidity: {ParamCode: model.WQMSParamTurbidity, DisplayName: "Kekeruhan (Turbidity)", ShortLabel: "Turbidity", Unit: "NTU", Rule: WQMSRuleMax, MaxLimit: ptrFloat(5.0), Label: "5 NTU"},
		model.WQMSParamFecalColi: {ParamCode: model.WQMSParamFecalColi, DisplayName: "Fecal Coliform", ShortLabel: "F.Coli", Unit: "MPN/100mL", Rule: WQMSRuleMax, MaxLimit: ptrFloat(100.0), Label: "100 MPN/100mL"},
	},
	model.WaterClass2: {
		model.WQMSParamPH:        {ParamCode: model.WQMSParamPH, DisplayName: "pH (Derajat Keasaman)", ShortLabel: "pH", Unit: "pH", Rule: WQMSRuleRange, MinLimit: ptrFloat(6.0), MaxLimit: ptrFloat(9.0), Label: "6.0 – 9.0"},
		model.WQMSParamDO:        {ParamCode: model.WQMSParamDO, DisplayName: "DO (Oksigen Terlarut)", ShortLabel: "DO", Unit: "mg/L", Rule: WQMSRuleInverseDO, MinLimit: ptrFloat(4.0), Label: "≥ 4.0 mg/L"},
		model.WQMSParamBOD:       {ParamCode: model.WQMSParamBOD, DisplayName: "BOD5", ShortLabel: "BOD", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(3.0), Label: "3.0 mg/L"},
		model.WQMSParamCOD:       {ParamCode: model.WQMSParamCOD, DisplayName: "COD", ShortLabel: "COD", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(25.0), Label: "25 mg/L"},
		model.WQMSParamTSS:       {ParamCode: model.WQMSParamTSS, DisplayName: "TSS (Padatan Tersuspensi)", ShortLabel: "TSS", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(50.0), Label: "50 mg/L"},
		model.WQMSParamNO3:       {ParamCode: model.WQMSParamNO3, DisplayName: "Nitrat (NO3-N)", ShortLabel: "NO3-N", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(10.0), Label: "10 mg/L"},
		model.WQMSParamNH3:       {ParamCode: model.WQMSParamNH3, DisplayName: "Amonia Total (NH3-N)", ShortLabel: "NH3-N", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(0.2), Label: "0.2 mg/L"},
		model.WQMSParamTDS:       {ParamCode: model.WQMSParamTDS, DisplayName: "TDS (Padatan Terlarut)", ShortLabel: "TDS", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(1000.0), Label: "1000 mg/L"},
		model.WQMSParamPO4:       {ParamCode: model.WQMSParamPO4, DisplayName: "Total Fosfat (PO4-P)", ShortLabel: "PO4-P", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(0.2), Label: "0.2 mg/L"},
		model.WQMSParamTurbidity: {ParamCode: model.WQMSParamTurbidity, DisplayName: "Kekeruhan (Turbidity)", ShortLabel: "Turbidity", Unit: "NTU", Rule: WQMSRuleMax, MaxLimit: ptrFloat(25.0), Label: "25 NTU"},
		model.WQMSParamFecalColi: {ParamCode: model.WQMSParamFecalColi, DisplayName: "Fecal Coliform", ShortLabel: "F.Coli", Unit: "MPN/100mL", Rule: WQMSRuleMax, MaxLimit: ptrFloat(1000.0), Label: "1000 MPN/100mL"},
	},
	model.WaterClass3: {
		model.WQMSParamPH:        {ParamCode: model.WQMSParamPH, DisplayName: "pH (Derajat Keasaman)", ShortLabel: "pH", Unit: "pH", Rule: WQMSRuleRange, MinLimit: ptrFloat(6.0), MaxLimit: ptrFloat(9.0), Label: "6.0 – 9.0"},
		model.WQMSParamDO:        {ParamCode: model.WQMSParamDO, DisplayName: "DO (Oksigen Terlarut)", ShortLabel: "DO", Unit: "mg/L", Rule: WQMSRuleInverseDO, MinLimit: ptrFloat(3.0), Label: "≥ 3.0 mg/L"},
		model.WQMSParamBOD:       {ParamCode: model.WQMSParamBOD, DisplayName: "BOD5", ShortLabel: "BOD", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(6.0), Label: "6.0 mg/L"},
		model.WQMSParamCOD:       {ParamCode: model.WQMSParamCOD, DisplayName: "COD", ShortLabel: "COD", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(40.0), Label: "40 mg/L"},
		model.WQMSParamTSS:       {ParamCode: model.WQMSParamTSS, DisplayName: "TSS (Padatan Tersuspensi)", ShortLabel: "TSS", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(100.0), Label: "100 mg/L"},
		model.WQMSParamNO3:       {ParamCode: model.WQMSParamNO3, DisplayName: "Nitrat (NO3-N)", ShortLabel: "NO3-N", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(20.0), Label: "20 mg/L"},
		model.WQMSParamNH3:       {ParamCode: model.WQMSParamNH3, DisplayName: "Amonia Total (NH3-N)", ShortLabel: "NH3-N", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(0.5), Label: "0.5 mg/L"},
		model.WQMSParamTDS:       {ParamCode: model.WQMSParamTDS, DisplayName: "TDS (Padatan Terlarut)", ShortLabel: "TDS", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(1000.0), Label: "1000 mg/L"},
		model.WQMSParamPO4:       {ParamCode: model.WQMSParamPO4, DisplayName: "Total Fosfat (PO4-P)", ShortLabel: "PO4-P", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(1.0), Label: "1.0 mg/L"},
		model.WQMSParamTurbidity: {ParamCode: model.WQMSParamTurbidity, DisplayName: "Kekeruhan (Turbidity)", ShortLabel: "Turbidity", Unit: "NTU", Rule: WQMSRuleMax, MaxLimit: ptrFloat(50.0), Label: "50 NTU"},
		model.WQMSParamFecalColi: {ParamCode: model.WQMSParamFecalColi, DisplayName: "Fecal Coliform", ShortLabel: "F.Coli", Unit: "MPN/100mL", Rule: WQMSRuleMax, MaxLimit: ptrFloat(2000.0), Label: "2000 MPN/100mL"},
	},
	model.WaterClass4: {
		model.WQMSParamPH:        {ParamCode: model.WQMSParamPH, DisplayName: "pH (Derajat Keasaman)", ShortLabel: "pH", Unit: "pH", Rule: WQMSRuleRange, MinLimit: ptrFloat(6.0), MaxLimit: ptrFloat(9.0), Label: "6.0 – 9.0"},
		model.WQMSParamDO:        {ParamCode: model.WQMSParamDO, DisplayName: "DO (Oksigen Terlarut)", ShortLabel: "DO", Unit: "mg/L", Rule: WQMSRuleInverseDO, MinLimit: ptrFloat(1.0), Label: "≥ 1.0 mg/L"},
		model.WQMSParamBOD:       {ParamCode: model.WQMSParamBOD, DisplayName: "BOD5", ShortLabel: "BOD", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(12.0), Label: "12 mg/L"},
		model.WQMSParamCOD:       {ParamCode: model.WQMSParamCOD, DisplayName: "COD", ShortLabel: "COD", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(80.0), Label: "80 mg/L"},
		model.WQMSParamTSS:       {ParamCode: model.WQMSParamTSS, DisplayName: "TSS (Padatan Tersuspensi)", ShortLabel: "TSS", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(400.0), Label: "400 mg/L"},
		model.WQMSParamNO3:       {ParamCode: model.WQMSParamNO3, DisplayName: "Nitrat (NO3-N)", ShortLabel: "NO3-N", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(20.0), Label: "20 mg/L"},
		model.WQMSParamNH3:       {ParamCode: model.WQMSParamNH3, DisplayName: "Amonia Total (NH3-N)", ShortLabel: "NH3-N", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(2.0), Label: "2.0 mg/L"},
		model.WQMSParamTDS:       {ParamCode: model.WQMSParamTDS, DisplayName: "TDS (Padatan Terlarut)", ShortLabel: "TDS", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(2000.0), Label: "2000 mg/L"},
		model.WQMSParamPO4:       {ParamCode: model.WQMSParamPO4, DisplayName: "Total Fosfat (PO4-P)", ShortLabel: "PO4-P", Unit: "mg/L", Rule: WQMSRuleMax, MaxLimit: ptrFloat(5.0), Label: "5.0 mg/L"},
		model.WQMSParamTurbidity: {ParamCode: model.WQMSParamTurbidity, DisplayName: "Kekeruhan (Turbidity)", ShortLabel: "Turbidity", Unit: "NTU", Rule: WQMSRuleMax, MaxLimit: ptrFloat(100.0), Label: "100 NTU"},
		model.WQMSParamFecalColi: {ParamCode: model.WQMSParamFecalColi, DisplayName: "Fecal Coliform", ShortLabel: "F.Coli", Unit: "MPN/100mL", Rule: WQMSRuleMax, MaxLimit: ptrFloat(2000.0), Label: "2000 MPN/100mL"},
	},
}

// NormalizeWaterClass normalizes a water class string to one of CLASS_1 .. CLASS_4 (default CLASS_2).
func NormalizeWaterClass(wc string) model.WQMSWaterClass {
	upper := strings.ToUpper(strings.TrimSpace(wc))
	switch upper {
	case "CLASS_1", "KELAS_1", "KELAS 1", "KELAS I", "1", "I":
		return model.WaterClass1
	case "CLASS_2", "KELAS_2", "KELAS 2", "KELAS II", "2", "II":
		return model.WaterClass2
	case "CLASS_3", "KELAS_3", "KELAS 3", "KELAS III", "3", "III":
		return model.WaterClass3
	case "CLASS_4", "KELAS_4", "KELAS 4", "KELAS IV", "4", "IV":
		return model.WaterClass4
	default:
		return model.WaterClass2
	}
}

// WQMSParameterInput holds input data for evaluating a single water quality parameter.
type WQMSParameterInput struct {
	ParamCode        model.WQMSParamCode
	ParameterID      *uint
	SourceType       string // SENSOR or LAB
	RawMeasured      float64
	SourceUnit       string
	ConversionFactor float64
	CustomLimitMin   *float64
	CustomLimitMax   *float64
	ValidSamples     int
	ExpectedSamples  int
	MinCompleteness  float64
}

// WQMSStationEvaluationOutput represents the deterministic result of evaluating station IPj.
type WQMSStationEvaluationOutput struct {
	WaterClass             model.WQMSWaterClass
	OverallIP              float64
	OverallCategory        model.WQMSIPCategory
	MaxRatio               float64
	AvgRatio               float64
	DominantParameter      model.WQMSParamCode
	DominantConcentration  float64
	DominantUnit           string
	TiedParameters         []model.WQMSParamCode
	CalculationStatus      model.ISPUStatus
	CompletenessPercent    float64
	ActiveParametersCount  int
	MissingParametersCount int
	ParameterResults       []model.WQMSParameterResult
	Notes                  string
}

// WQMSCalculator executes deterministic Kepmen LH No. 115/2003 (IPj) and Permen LHK No. 27/2021 (IKA) calculations.
type WQMSCalculator struct{}

// NewWQMSCalculator instantiates a WQMSCalculator.
func NewWQMSCalculator() *WQMSCalculator {
	return &WQMSCalculator{}
}

// ClassifyIP returns the official Kepmen LH No. 115 Tahun 2003 Pollution Index category.
func ClassifyIP(ip float64) model.WQMSIPCategory {
	if math.IsNaN(ip) || math.IsInf(ip, 0) || ip < 0 {
		return model.WQMSCategoryUnavailable
	}
	if ip <= 1.0 {
		return model.WQMSCategoryMemenuhi
	}
	if ip <= 5.0 {
		return model.WQMSCategoryCemarRingan
	}
	if ip <= 10.0 {
		return model.WQMSCategoryCemarSedang
	}
	return model.WQMSCategoryCemarBerat
}

// ClassifyIKA returns the official Permen LHK No. 27 Tahun 2021 Water Quality Index category & Indonesian label.
func ClassifyIKA(score float64) (model.WQMSIKACategory, string) {
	if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 {
		return model.IKACategoryUnavailable, "Tidak Tersedia"
	}
	if score >= 90.0 {
		return model.IKACategorySangatBaik, "Sangat Baik"
	}
	if score >= 70.0 {
		return model.IKACategoryBaik, "Baik"
	}
	if score >= 50.0 {
		return model.IKACategorySedang, "Sedang"
	}
	if score >= 25.0 {
		return model.IKACategoryKurang, "Kurang"
	}
	return model.IKACategorySangatKurang, "Sangat Kurang"
}

// EvaluateParameter computes (Ci/Lij)awal and logarithmic (Ci/Lij)baru for one parameter.
func (c *WQMSCalculator) EvaluateParameter(in WQMSParameterInput, waterClass model.WQMSWaterClass, doSaturationMgL float64) (model.WQMSParameterResult, bool) {
	classStandards, ok := OfficialWQMSStandards[waterClass]
	if !ok {
		waterClass = model.WaterClass2
		classStandards = OfficialWQMSStandards[model.WaterClass2]
	}
	std, hasStd := classStandards[in.ParamCode]
	if !hasStd {
		return model.WQMSParameterResult{}, false
	}

	if doSaturationMgL <= 0 {
		doSaturationMgL = 8.0
	}

	factor := in.ConversionFactor
	if factor <= 0 {
		factor = 1.0
	}
	ci := in.RawMeasured * factor
	if math.IsNaN(ci) || math.IsInf(ci, 0) {
		return model.WQMSParameterResult{}, false
	}

	minLim := std.MinLimit
	maxLim := std.MaxLimit
	label := std.Label
	if in.CustomLimitMin != nil {
		minLim = in.CustomLimitMin
	}
	if in.CustomLimitMax != nil {
		maxLim = in.CustomLimitMax
	}
	if in.CustomLimitMin != nil || in.CustomLimitMax != nil {
		switch std.Rule {
		case WQMSRuleRange:
			if minLim != nil && maxLim != nil {
				label = fmt.Sprintf("%.2g – %.2g", *minLim, *maxLim)
			}
		case WQMSRuleInverseDO:
			if minLim != nil {
				label = fmt.Sprintf("≥ %.2g %s", *minLim, std.Unit)
			}
		default:
			if maxLim != nil {
				label = fmt.Sprintf("%.2g %s", *maxLim, std.Unit)
			}
		}
	}

	var rawRatio float64
	var formulaNote string

	switch std.Rule {
	case WQMSRuleRange:
		lMin := 6.0
		lMax := 9.0
		if minLim != nil {
			lMin = *minLim
		}
		if maxLim != nil {
			lMax = *maxLim
		}
		lAvg := (lMin + lMax) / 2.0
		if ci < lAvg && (lMin-lAvg) != 0 {
			rawRatio = (ci - lAvg) / (lMin - lAvg)
			formulaNote = fmt.Sprintf("pH < %.2f: (%.2f - %.2f) / (%.2f - %.2f) = %.4f", lAvg, ci, lAvg, lMin, lAvg, rawRatio)
		} else if (lMax - lAvg) != 0 {
			rawRatio = (ci - lAvg) / (lMax - lAvg)
			formulaNote = fmt.Sprintf("pH ≥ %.2f: (%.2f - %.2f) / (%.2f - %.2f) = %.4f", lAvg, ci, lAvg, lMax, lAvg, rawRatio)
		}
	case WQMSRuleInverseDO:
		lMin := 4.0
		if minLim != nil {
			lMin = *minLim
		}
		cim := doSaturationMgL
		if cim <= lMin {
			cim = lMin + 2.0
		}
		if ci >= cim {
			rawRatio = 0.0
			formulaNote = fmt.Sprintf("DO %.2f ≥ Cim %.2f: defisit oksigen = 0.00", ci, cim)
		} else {
			if ci < 0 {
				ci = 0
			}
			rawRatio = (cim - ci) / (cim - lMin)
			formulaNote = fmt.Sprintf("Defisit DO: (%.2f - %.2f) / (%.2f - %.2f) = %.4f", cim, ci, cim, lMin, rawRatio)
		}
	default: // WQMSRuleMax
		lMax := 1.0
		if maxLim != nil && *maxLim > 0 {
			lMax = *maxLim
		}
		if ci < 0 {
			ci = 0
		}
		rawRatio = ci / lMax
		formulaNote = fmt.Sprintf("Ci / Lij: %.2f / %.2f = %.4f", ci, lMax, rawRatio)
	}

	if rawRatio < 0 {
		rawRatio = 0
	}

	exceeds := rawRatio > 1.0
	adjustedRatio := rawRatio
	logCorrected := false
	if rawRatio > 1.0 {
		adjustedRatio = 1.0 + 5.0*math.Log10(rawRatio)
		logCorrected = true
		formulaNote = fmt.Sprintf("%s → Koreksi log: 1.0 + 5·log10(%.4f) = %.4f", formulaNote, rawRatio, adjustedRatio)
	}

	completeness := 100.0
	status := model.ISPUStatusValid
	sourceType := strings.ToUpper(strings.TrimSpace(in.SourceType))
	if sourceType == "" {
		sourceType = "SENSOR"
	}

	if sourceType == "SENSOR" {
		if in.ExpectedSamples > 0 {
			completeness = (float64(in.ValidSamples) / float64(in.ExpectedSamples)) * 100.0
			if completeness > 100.0 {
				completeness = 100.0
			}
		} else if in.ValidSamples == 0 {
			completeness = 0
		}
		minComp := in.MinCompleteness
		if minComp <= 0 {
			minComp = 75.0
		}
		if in.ValidSamples <= 0 {
			status = model.ISPUStatusInvalid
		} else if completeness < minComp {
			status = model.ISPUStatusProvisional
		}
	}

	srcUnit := in.SourceUnit
	if srcUnit == "" {
		srcUnit = std.Unit
	}

	res := model.WQMSParameterResult{
		ParamCode:           in.ParamCode,
		ParameterID:         in.ParameterID,
		SourceType:          sourceType,
		RawMeasured:         roundTo(in.RawMeasured, 4),
		SourceUnit:          srcUnit,
		NormalizedMeasured:  roundTo(ci, 4),
		NormalizedUnit:      std.Unit,
		StandardRule:        std.Rule,
		StandardMin:         minLim,
		StandardMax:         maxLim,
		StandardLabel:       label,
		RawRatio:            roundTo(rawRatio, 4),
		AdjustedRatio:       roundTo(adjustedRatio, 4),
		LogCorrected:        logCorrected,
		ExceedsStandard:     exceeds,
		CompletenessPercent: roundTo(completeness, 2),
		ValidSamplesCount:   in.ValidSamples,
		ExpectedSamples:     in.ExpectedSamples,
		Status:              status,
		FormulaNote:         formulaNote,
	}
	return res, true
}

// EvaluateStation computes the overall station Pollution Index (IPj) from evaluated parameter inputs.
func (c *WQMSCalculator) EvaluateStation(inputs []WQMSParameterInput, waterClass model.WQMSWaterClass, doSaturationMgL float64) WQMSStationEvaluationOutput {
	wc := NormalizeWaterClass(string(waterClass))
	out := WQMSStationEvaluationOutput{
		WaterClass:        wc,
		OverallCategory:   model.WQMSCategoryUnavailable,
		CalculationStatus: model.ISPUStatusInvalid,
	}

	if len(inputs) == 0 {
		out.Notes = "Belum ada parameter air yang dipetakan atau memiliki data valid."
		return out
	}

	var results []model.WQMSParameterResult
	var sumRatio float64
	var maxRatio float64 = -1.0
	var sumCompleteness float64
	var sensorCount int
	var provisionalCount int
	var missingCount int

	for _, in := range inputs {
		if strings.ToUpper(in.SourceType) != "LAB" && in.ValidSamples <= 0 {
			missingCount++
			continue
		}
		pr, ok := c.EvaluateParameter(in, wc, doSaturationMgL)
		if !ok {
			continue
		}
		results = append(results, pr)
		sumRatio += pr.AdjustedRatio
		if pr.AdjustedRatio > maxRatio {
			maxRatio = pr.AdjustedRatio
		}
		if pr.SourceType == "SENSOR" {
			sumCompleteness += pr.CompletenessPercent
			sensorCount++
			if pr.Status == model.ISPUStatusProvisional {
				provisionalCount++
			}
		}
	}

	if len(results) == 0 {
		out.MissingParametersCount = missingCount
		out.Notes = "Tidak ada sampel parameter valid pada jendela waktu evaluasi."
		return out
	}

	avgRatio := sumRatio / float64(len(results))
	ipVal := math.Sqrt((maxRatio*maxRatio + avgRatio*avgRatio) / 2.0)

	// Identify dominant & tied parameters following official priority order
	priorityIndex := make(map[model.WQMSParamCode]int, len(model.AllOfficialWQMSParameters))
	for idx, code := range model.AllOfficialWQMSParameters {
		priorityIndex[code] = idx
	}

	var tied []model.WQMSParamCode
	dominantIdx := -1
	for i := range results {
		if math.Abs(results[i].AdjustedRatio-maxRatio) < 1e-6 {
			tied = append(tied, results[i].ParamCode)
			if dominantIdx == -1 {
				dominantIdx = i
			} else {
				pCur, okCur := priorityIndex[results[i].ParamCode]
				pBest, okBest := priorityIndex[results[dominantIdx].ParamCode]
				if okCur && (!okBest || pCur < pBest) {
					dominantIdx = i
				}
			}
		}
	}
	if dominantIdx >= 0 {
		results[dominantIdx].IsDominant = true
		out.DominantParameter = results[dominantIdx].ParamCode
		out.DominantConcentration = results[dominantIdx].NormalizedMeasured
		out.DominantUnit = results[dominantIdx].NormalizedUnit
	}

	compPct := 100.0
	if sensorCount > 0 {
		compPct = sumCompleteness / float64(sensorCount)
	}

	calcStatus := model.ISPUStatusValid
	if missingCount > 0 && provisionalCount > 0 {
		calcStatus = model.ISPUStatusProvisional
	} else if provisionalCount > 0 {
		calcStatus = model.ISPUStatusProvisional
	} else if missingCount > 0 {
		calcStatus = model.ISPUStatusIncomplete
	}

	out.OverallIP = roundTo(ipVal, 2)
	out.OverallCategory = ClassifyIP(out.OverallIP)
	out.MaxRatio = roundTo(maxRatio, 2)
	out.AvgRatio = roundTo(avgRatio, 2)
	out.TiedParameters = tied
	out.CalculationStatus = calcStatus
	out.CompletenessPercent = roundTo(compPct, 1)
	out.ActiveParametersCount = len(results)
	out.MissingParametersCount = missingCount
	out.ParameterResults = results
	out.Notes = fmt.Sprintf("IPj = √(((%.2f)² + (%.2f)²) / 2) = %.2f (%s)", out.MaxRatio, out.AvgRatio, out.OverallIP, out.OverallCategory)
	return out
}

// ComputeIKASummary computes the Permen LHK No. 27 Tahun 2021 Water Quality Index (IKA) from a slice of IP categories.
func (c *WQMSCalculator) ComputeIKASummary(categories []model.WQMSIPCategory, waterClass model.WQMSWaterClass, windowDays int, source string) model.WQMSIKASummary {
	if windowDays <= 0 {
		windowDays = 30
	}
	summary := model.WQMSIKASummary{
		WaterClass:       NormalizeWaterClass(string(waterClass)),
		WindowDays:       windowDays,
		EvaluationSource: source,
		Category:         model.IKACategoryUnavailable,
		CategoryLabel:    "Tidak Tersedia",
		StandardRef:      "Permen LHK No. 27 Tahun 2021",
	}

	validTotal := 0
	for _, cat := range categories {
		switch cat {
		case model.WQMSCategoryMemenuhi:
			summary.CountMemenuhi++
			validTotal++
		case model.WQMSCategoryCemarRingan:
			summary.CountCemarRingan++
			validTotal++
		case model.WQMSCategoryCemarSedang:
			summary.CountCemarSedang++
			validTotal++
		case model.WQMSCategoryCemarBerat:
			summary.CountCemarBerat++
			validTotal++
		}
	}

	summary.TotalPeriods = validTotal
	if validTotal == 0 {
		return summary
	}

	fBaik := float64(summary.CountMemenuhi) / float64(validTotal)
	fRingan := float64(summary.CountCemarRingan) / float64(validTotal)
	fSedang := float64(summary.CountCemarSedang) / float64(validTotal)
	fBerat := float64(summary.CountCemarBerat) / float64(validTotal)

	score := (fBaik * 70.0) + (fRingan * 50.0) + (fSedang * 30.0) + (fBerat * 10.0)
	cat, label := ClassifyIKA(score)

	summary.Score = roundTo(score, 1)
	summary.Category = cat
	summary.CategoryLabel = label
	summary.PctMemenuhi = roundTo(fBaik*100.0, 1)
	summary.PctCemarRingan = roundTo(fRingan*100.0, 1)
	summary.PctCemarSedang = roundTo(fSedang*100.0, 1)
	summary.PctCemarBerat = roundTo(fBerat*100.0, 1)
	return summary
}

// DetectOfficialWQMSParameter inspects a sensor's code, name, and unit to auto-match an official WQMS parameter.
// Explicitly avoids matching meteorological/gas AQMS sensors (e.g. CO, SO2, NO2, PM10, PM2.5, Wind, Air Temp).
func DetectOfficialWQMSParameter(paramCode, paramName, unit string) (model.WQMSParamCode, string, float64) {
	code := strings.ToUpper(strings.TrimSpace(paramCode))
	name := strings.ToUpper(strings.TrimSpace(paramName))
	u := strings.TrimSpace(unit)

	// Guard against AQMS ambient gas/particulate parameters
	if code == "CO" || code == "SO2" || code == "NO2" || code == "O3" || code == "HC" || code == "NMHC" ||
		strings.Contains(code, "PM10") || strings.Contains(code, "PM25") || strings.Contains(code, "PM2.5") ||
		strings.Contains(name, "PARTICULATE") || strings.Contains(name, "KARBON MONOKSIDA") {
		return "", "", 1.0
	}

	switch {
	case code == "PH" || code == "W_PH" || code == "PH_AIR" || name == "PH" || strings.Contains(name, "PH AIR") || strings.Contains(name, "DERAJAT KEASAMAN"):
		if u == "" {
			u = "pH"
		}
		return model.WQMSParamPH, u, 1.0

	case code == "DO" || code == "W_DO" || strings.Contains(name, "DISSOLVED OXYGEN") || strings.Contains(name, "OKSIGEN TERLARUT"):
		if u == "" {
			u = "mg/L"
		}
		return model.WQMSParamDO, u, 1.0

	case code == "COD" || code == "W_COD" || strings.Contains(name, "CHEMICAL OXYGEN DEMAND") || name == "COD":
		if u == "" {
			u = "mg/L"
		}
		return model.WQMSParamCOD, u, 1.0

	case code == "BOD" || code == "BOD5" || code == "W_BOD" || strings.Contains(name, "BIOCHEMICAL OXYGEN DEMAND") || strings.Contains(name, "BOD"):
		if u == "" {
			u = "mg/L"
		}
		return model.WQMSParamBOD, u, 1.0

	case code == "TSS" || code == "W_TSS" || strings.Contains(name, "TOTAL SUSPENDED") || strings.Contains(name, "PADATAN TERSUSPENSI"):
		if u == "" {
			u = "mg/L"
		}
		return model.WQMSParamTSS, u, 1.0

	case code == "NO3" || code == "NO3_N" || code == "NO3-N" || code == "NITRAT" || code == "NITRATE" || strings.Contains(name, "NITRAT") || strings.Contains(name, "NITRATE"):
		if u == "" {
			u = "mg/L"
		}
		return model.WQMSParamNO3, u, 1.0

	case code == "NH3" || code == "NH3_N" || code == "NH3-N" || code == "NH4" || code == "AMONIA" || code == "AMMONIA" || strings.Contains(name, "AMONIA") || strings.Contains(name, "AMMONIA"):
		if u == "" {
			u = "mg/L"
		}
		return model.WQMSParamNH3, u, 1.0

	case code == "TDS" || code == "W_TDS" || strings.Contains(name, "TOTAL DISSOLVED") || strings.Contains(name, "PADATAN TERLARUT"):
		if u == "" {
			u = "mg/L"
		}
		return model.WQMSParamTDS, u, 1.0

	case code == "PO4" || code == "PO4_P" || code == "PO4-P" || code == "FOSFAT" || strings.Contains(name, "FOSFAT") || strings.Contains(name, "PHOSPHATE"):
		if u == "" {
			u = "mg/L"
		}
		return model.WQMSParamPO4, u, 1.0

	case code == "TURB" || code == "TURBIDITY" || code == "KEKERUHAN" || strings.Contains(name, "TURBIDITY") || strings.Contains(name, "KEKERUHAN"):
		if u == "" {
			u = "NTU"
		}
		return model.WQMSParamTurbidity, u, 1.0

	case code == "FECAL_COLI" || code == "COLIFORM" || strings.Contains(name, "COLIFORM"):
		if u == "" {
			u = "MPN/100mL"
		}
		return model.WQMSParamFecalColi, u, 1.0
	}

	return "", "", 1.0
}

func roundTo(val float64, decimals int) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0
	}
	pow := math.Pow(10, float64(decimals))
	return math.Round(val*pow) / pow
}

