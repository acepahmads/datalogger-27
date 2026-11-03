package model

import (
	"time"
)

// WQMSWaterClass represents the official river/surface water quality class under Lampiran VI PP No. 22 Tahun 2021.
type WQMSWaterClass string

const (
	WaterClass1 WQMSWaterClass = "CLASS_1" // Kelas I — Air baku air minum
	WaterClass2 WQMSWaterClass = "CLASS_2" // Kelas II — Prasarana/sarana rekreasi air, budidaya ikan air tawar, peternakan, pengairan
	WaterClass3 WQMSWaterClass = "CLASS_3" // Kelas III — Pembudidayaan ikan air tawar, peternakan, pengairan pertanaman
	WaterClass4 WQMSWaterClass = "CLASS_4" // Kelas IV — Pengairan pertanaman
)

// AllWQMSWaterClasses lists supported PP No. 22 Tahun 2021 surface water classes.
var AllWQMSWaterClasses = []WQMSWaterClass{
	WaterClass1,
	WaterClass2,
	WaterClass3,
	WaterClass4,
}

// WQMSParamCode represents an official water quality parameter evaluated in IP & IKA calculations.
type WQMSParamCode string

const (
	WQMSParamPH        WQMSParamCode = "PH"
	WQMSParamDO        WQMSParamCode = "DO"
	WQMSParamBOD       WQMSParamCode = "BOD"
	WQMSParamCOD       WQMSParamCode = "COD"
	WQMSParamTSS       WQMSParamCode = "TSS"
	WQMSParamNO3       WQMSParamCode = "NO3"
	WQMSParamNH3       WQMSParamCode = "NH3"
	WQMSParamTDS       WQMSParamCode = "TDS"
	WQMSParamPO4       WQMSParamCode = "PO4"
	WQMSParamTurbidity WQMSParamCode = "TURBIDITY"
	WQMSParamFecalColi WQMSParamCode = "FECAL_COLI"
)

// AllOfficialWQMSParameters lists supported water quality parameters in tie-breaking priority order.
var AllOfficialWQMSParameters = []WQMSParamCode{
	WQMSParamCOD,
	WQMSParamBOD,
	WQMSParamDO,
	WQMSParamTSS,
	WQMSParamPH,
	WQMSParamNH3,
	WQMSParamNO3,
	WQMSParamPO4,
	WQMSParamTDS,
	WQMSParamTurbidity,
	WQMSParamFecalColi,
}

// WQMSIPCategory represents the official Pollution Index (Indeks Pencemaran) status under Kepmen LH No. 115 Tahun 2003.
type WQMSIPCategory string

const (
	WQMSCategoryMemenuhi    WQMSIPCategory = "MEMENUHI_BAKU_MUTU" // 0 <= IP <= 1.0
	WQMSCategoryCemarRingan WQMSIPCategory = "CEMAR_RINGAN"       // 1.0 < IP <= 5.0
	WQMSCategoryCemarSedang WQMSIPCategory = "CEMAR_SEDANG"       // 5.0 < IP <= 10.0
	WQMSCategoryCemarBerat  WQMSIPCategory = "CEMAR_BERAT"        // IP > 10.0
	WQMSCategoryUnavailable WQMSIPCategory = "TIDAK_TERSEDIA"
)

// WQMSIKACategory represents the 5-tier Water Quality Index (Indeks Kualitas Air) category under Permen LHK No. 27 Tahun 2021.
type WQMSIKACategory string

const (
	IKACategorySangatBaik   WQMSIKACategory = "SANGAT_BAIK"   // 90 <= IKA <= 100
	IKACategoryBaik         WQMSIKACategory = "BAIK"          // 70 <= IKA < 90
	IKACategorySedang       WQMSIKACategory = "SEDANG"        // 50 <= IKA < 70
	IKACategoryKurang       WQMSIKACategory = "KURANG"        // 25 <= IKA < 50
	IKACategorySangatKurang WQMSIKACategory = "SANGAT_KURANG" // 0 <= IKA < 25
	IKACategoryUnavailable  WQMSIKACategory = "TIDAK_TERSEDIA"
)

// WQMSStationConfig stores station-level configuration for WQMS IP & IKA evaluation.
type WQMSStationConfig struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	DeviceID        uint           `gorm:"uniqueIndex;not null;default:0" json:"device_id"` // 0 = unified station
	WaterClass      WQMSWaterClass `gorm:"size:16;not null;default:'CLASS_2'" json:"water_class"`
	IKAWindowDays   int            `gorm:"not null;default:30" json:"ika_window_days"`
	DOSaturationMgL float64        `gorm:"not null;default:8.0" json:"do_saturation_mgl"` // Cim theoretical saturation DO (mg/L)
	MinCompleteness float64        `gorm:"not null;default:75.0" json:"min_completeness"`
	UpdatedBy       string         `gorm:"size:64" json:"updated_by,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// WQMSParameterMapping maps either a live datalogger parameter (SENSOR) or periodic laboratory value (LAB)
// to an official WQMS water quality parameter.
type WQMSParameterMapping struct {
	ID               uint          `gorm:"primaryKey" json:"id"`
	DeviceID         uint          `gorm:"uniqueIndex:idx_wqms_dev_param;index;not null;default:0" json:"device_id"`
	Device           *Device       `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	ParameterID      *uint         `gorm:"index" json:"parameter_id"`
	Parameter        *Parameter    `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`
	ParamCode        WQMSParamCode `gorm:"size:24;uniqueIndex:idx_wqms_dev_param;not null" json:"param_code"`
	SourceType       string        `gorm:"size:16;not null;default:'SENSOR'" json:"source_type"` // SENSOR or LAB
	LabValue         *float64      `json:"lab_value,omitempty"`
	LabSampledAt     *time.Time    `json:"lab_sampled_at,omitempty"`
	SourceUnit       string        `gorm:"size:32;not null;default:'mg/L'" json:"source_unit"`
	NormalizedUnit   string        `gorm:"size:32;not null;default:'mg/L'" json:"normalized_unit"`
	ConversionFactor float64       `gorm:"default:1.0" json:"conversion_factor"`
	CustomLimitMin   *float64      `json:"custom_limit_min,omitempty"`
	CustomLimitMax   *float64      `json:"custom_limit_max,omitempty"`
	MinCompleteness  float64       `gorm:"default:75.0" json:"min_completeness"`
	Enabled          bool          `gorm:"default:true" json:"enabled"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// WQMSCalculationResult stores a persisted Pollution Index (IPj) evaluation for a station, period, and mode.
type WQMSCalculationResult struct {
	ID                     uint                  `gorm:"primaryKey" json:"id"`
	DeviceID               uint                  `gorm:"uniqueIndex:idx_wqms_dev_period_mode;index;not null;default:0" json:"device_id"`
	Device                 *Device               `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	CalculationMode        ISPUCalculationMode   `gorm:"size:32;default:'REGULATORY';uniqueIndex:idx_wqms_dev_period_mode;index;not null" json:"calculation_mode"`
	WaterClass             WQMSWaterClass        `gorm:"size:16;not null;default:'CLASS_2'" json:"water_class"`
	PeriodStart            time.Time             `gorm:"uniqueIndex:idx_wqms_dev_period_mode;index;not null" json:"period_start"`
	PeriodEnd              time.Time             `gorm:"uniqueIndex:idx_wqms_dev_period_mode;index;not null" json:"period_end"`
	CalculatedAt           time.Time             `gorm:"not null" json:"calculated_at"`
	OverallIP              float64               `gorm:"not null" json:"overall_ip"`
	OverallCategory        WQMSIPCategory        `gorm:"size:32;not null" json:"overall_category"`
	MaxRatio               float64               `gorm:"not null" json:"max_ratio"` // (Ci/Lij)M
	AvgRatio               float64               `gorm:"not null" json:"avg_ratio"` // (Ci/Lij)R
	DominantParameter      WQMSParamCode         `gorm:"size:24;not null" json:"dominant_parameter"`
	DominantConcentration  float64               `json:"dominant_concentration"`
	DominantUnit           string                `gorm:"size:32" json:"dominant_unit"`
	TiedParameters         string                `gorm:"size:128" json:"tied_parameters,omitempty"`
	CalculationStatus      ISPUStatus            `gorm:"size:32;not null" json:"calculation_status"`
	CompletenessPercent    float64               `gorm:"not null" json:"completeness_percent"`
	ActiveParametersCount  int                   `gorm:"default:0" json:"active_parameters_count"`
	MissingParametersCount int                   `gorm:"default:0" json:"missing_parameters_count"`
	Method                 string                `gorm:"size:96;default:'KEPMEN_LH_115_2003_PP_22_2021'" json:"method"`
	CalculationReason      string                `gorm:"size:64;default:'AGGREGATION_COMPLETED'" json:"calculation_reason"`
	Notes                  string                `gorm:"type:text" json:"notes,omitempty"`
	ParameterResults       []WQMSParameterResult `gorm:"foreignKey:CalculationResultID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"parameter_results,omitempty"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}

// WQMSParameterResult stores per-parameter Ci/Lij ratio evaluation details for a WQMS calculation.
type WQMSParameterResult struct {
	ID                  uint          `gorm:"primaryKey" json:"id"`
	CalculationResultID uint          `gorm:"index;not null" json:"calculation_result_id"`
	ParamCode           WQMSParamCode `gorm:"size:24;index;not null" json:"param_code"`
	ParameterID         *uint         `gorm:"index" json:"parameter_id"`
	Parameter           *Parameter    `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`
	SourceType          string        `gorm:"size:16;default:'SENSOR'" json:"source_type"`
	RawMeasured         float64       `json:"raw_measured"`
	SourceUnit          string        `gorm:"size:32" json:"source_unit"`
	NormalizedMeasured  float64       `json:"normalized_measured"` // Ci
	NormalizedUnit      string        `gorm:"size:32" json:"normalized_unit"`
	StandardRule        string        `gorm:"size:24" json:"standard_rule"` // MAX, RANGE, INVERSE_DO
	StandardMin         *float64      `json:"standard_min,omitempty"`
	StandardMax         *float64      `json:"standard_max,omitempty"`
	StandardLabel       string        `gorm:"size:64" json:"standard_label"` // e.g., "6.0–9.0", "≥4.0", "25"
	RawRatio            float64       `json:"raw_ratio"`                     // (Ci/Lij) awal
	AdjustedRatio       float64       `json:"adjusted_ratio"`                // (Ci/Lij) baru (log-corrected if > 1.0)
	LogCorrected        bool          `json:"log_corrected"`
	ExceedsStandard     bool          `json:"exceeds_standard"`
	IsDominant          bool          `gorm:"default:false" json:"is_dominant"`
	CompletenessPercent float64       `json:"completeness_percent"`
	ValidSamplesCount   int           `json:"valid_samples_count"`
	ExpectedSamples     int           `json:"expected_samples"`
	Status              ISPUStatus    `gorm:"size:32" json:"status"`
	FormulaNote         string        `gorm:"size:255" json:"formula_note,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
}

// WQMSIKASummary holds the aggregated Water Quality Index (Indeks Kualitas Air) score & status breakdown
// under Permen LHK No. 27 Tahun 2021.
type WQMSIKASummary struct {
	Score            float64         `json:"score"` // 0 - 100
	Category         WQMSIKACategory `json:"category"`
	CategoryLabel    string          `json:"category_label"`
	WaterClass       WQMSWaterClass  `json:"water_class"`
	WindowDays       int             `json:"window_days"`
	TotalPeriods     int             `json:"total_periods"`
	EvaluationSource string          `json:"evaluation_source"` // REGULATORY_24H or OPERATIONAL_LIVE
	CountMemenuhi    int             `json:"count_memenuhi"`
	PctMemenuhi      float64         `json:"pct_memenuhi"`
	CountCemarRingan int             `json:"count_cemar_ringan"`
	PctCemarRingan   float64         `json:"pct_cemar_ringan"`
	CountCemarSedang int             `json:"count_cemar_sedang"`
	PctCemarSedang   float64         `json:"pct_cemar_sedang"`
	CountCemarBerat  int             `json:"count_cemar_berat"`
	PctCemarBerat    float64         `json:"pct_cemar_berat"`
	StandardRef      string          `json:"standard_ref"`
}

