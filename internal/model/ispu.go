package model

import (
	"time"
)

// ISPUParameter represents one of the 7 official pollutants regulated under Permen LHK No. 14 Tahun 2020.
type ISPUParameter string

const (
	PollutantPM10 ISPUParameter = "PM10"
	PollutantPM25 ISPUParameter = "PM25"
	PollutantSO2  ISPUParameter = "SO2"
	PollutantCO   ISPUParameter = "CO"
	PollutantO3   ISPUParameter = "O3"
	PollutantNO2  ISPUParameter = "NO2"
	PollutantHC   ISPUParameter = "HC"
)

// AllOfficialPollutants lists the seven official pollutants in priority tie-breaking order.
var AllOfficialPollutants = []ISPUParameter{
	PollutantPM25,
	PollutantPM10,
	PollutantO3,
	PollutantNO2,
	PollutantSO2,
	PollutantCO,
	PollutantHC,
}

// ISPUCategory represents the official ISPU category defined in Lampiran II Permen LHK No. 14 Tahun 2020.
type ISPUCategory string

const (
	CategoryBaik             ISPUCategory = "BAIK"
	CategorySedang           ISPUCategory = "SEDANG"
	CategoryTidakSehat       ISPUCategory = "TIDAK_SEHAT"
	CategorySangatTidakSehat ISPUCategory = "SANGAT_TIDAK_SEHAT"
	CategoryBerbahaya        ISPUCategory = "BERBAHAYA"
	CategoryUnavailable      ISPUCategory = "TIDAK_TERSEDIA"
)

// ISPUStatus represents the regulatory calculation validity state.
type ISPUStatus string

const (
	ISPUStatusValid       ISPUStatus = "VALID"       // Meets >=75% temporal data completeness and quality
	ISPUStatusProvisional ISPUStatus = "PROVISIONAL" // In-progress window or completeness >0% and <75%
	ISPUStatusIncomplete  ISPUStatus = "INCOMPLETE"  // Some parameters missing
	ISPUStatusInvalid     ISPUStatus = "INVALID"     // 0% data completeness or hardware/data error
)

// ISPUParameterMapping maps a physical/virtual datalogger parameter to an official ISPU pollutant.
type ISPUParameterMapping struct {
	ID               uint          `gorm:"primaryKey" json:"id"`
	DeviceID         uint          `gorm:"uniqueIndex:idx_device_pollutant;index;not null" json:"device_id"`
	Device           *Device       `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	ParameterID      uint          `gorm:"index;not null" json:"parameter_id"`
	Parameter        *Parameter    `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`
	PollutantCode    ISPUParameter `gorm:"size:16;uniqueIndex:idx_device_pollutant;not null" json:"pollutant_code"`
	SourceUnit       string        `gorm:"size:32;not null;default:'µg/m³'" json:"source_unit"`
	NormalizedUnit   string        `gorm:"size:32;not null;default:'µg/m³'" json:"normalized_unit"`
	ConversionFactor float64       `gorm:"default:1.0" json:"conversion_factor"`
	ConversionRule   string        `gorm:"size:64;default:'DIRECT'" json:"conversion_rule"` // DIRECT, PPM_TO_UGM3, PPB_TO_UGM3, MG_TO_UGM3
	MinCompleteness  float64       `gorm:"default:75.0" json:"min_completeness"`           // Permen LHK 14/2020: 75%
	DailyStartHour   int           `gorm:"default:0" json:"daily_start_hour"`               // Jam mulai hitung harian (0-23, misal 00:00 atau 09:00)
	Enabled          bool          `gorm:"default:true" json:"enabled"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// ISPUCalculationMode defines whether the calculation is Mode A (Operational per-aggregation) or Mode B (Regulatory 24h).
type ISPUCalculationMode string

const (
	CalculationModeRegulatory  ISPUCalculationMode = "REGULATORY"
	CalculationModeOperational ISPUCalculationMode = "OPERATIONAL"
)

// ISPUCalculationResult stores the persisted calculation result for a station, period, and mode.
type ISPUCalculationResult struct {
	ID                         uint                  `gorm:"primaryKey" json:"id"`
	DeviceID                   uint                  `gorm:"uniqueIndex:idx_dev_period_mode;index;not null" json:"device_id"`
	Device                     *Device               `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	CalculationMode            ISPUCalculationMode   `gorm:"size:32;default:'REGULATORY';uniqueIndex:idx_dev_period_mode;index;not null" json:"calculation_mode"`
	PeriodStart                time.Time             `gorm:"uniqueIndex:idx_dev_period_mode;index;not null" json:"period_start"`
	PeriodEnd                  time.Time             `gorm:"uniqueIndex:idx_dev_period_mode;index;not null" json:"period_end"`
	CalculatedAt               time.Time             `gorm:"not null" json:"calculated_at"`
	OverallISPU                float64               `gorm:"not null" json:"overall_ispu"`
	OverallCategory            ISPUCategory          `gorm:"size:32;not null" json:"overall_category"`
	DominantPollutant          ISPUParameter         `gorm:"size:16;not null" json:"dominant_pollutant"`
	TiedPollutants             string                `gorm:"size:64" json:"tied_pollutants,omitempty"`
	CompletenessPercent        float64               `gorm:"not null;default:0" json:"completeness_percent"`
	CalculationStatus          ISPUStatus            `gorm:"size:32;not null" json:"calculation_status"`
	Method                     string                `gorm:"size:64;default:'PERMEN_LHK_14_2020'" json:"method"`
	Timezone                   string                `gorm:"size:64;default:'Asia/Jakarta'" json:"timezone"`
	AggregationISPU            float64               `gorm:"default:0" json:"aggregation_ispu"`
	AggregationCategory        ISPUCategory          `gorm:"size:32;default:'TIDAK_TERSEDIA'" json:"aggregation_category"`
	AggregationDominant        ISPUParameter         `gorm:"size:16" json:"aggregation_dominant,omitempty"`
	AggregationIntervalSeconds int                   `gorm:"default:120" json:"aggregation_interval_seconds"`
	AggregationFunction        string                `gorm:"size:32;default:'AVG'" json:"aggregation_function"`
	SourceBucketIdentifier     string                `gorm:"size:32;index" json:"source_bucket_identifier"`
	TriggerReason              string                `gorm:"size:64;default:'AGGREGATION_COMPLETED'" json:"trigger_reason"`
	ParameterResults           []ISPUParameterResult `gorm:"foreignKey:CalculationResultID;constraint:OnDelete:CASCADE" json:"parameter_results,omitempty"`
	Notes                      string                `gorm:"size:255" json:"notes,omitempty"`
	CreatedAt                  time.Time             `json:"created_at"`
	UpdatedAt                  time.Time             `json:"updated_at"`
}

// ISPUParameterResult stores the per-pollutant calculation breakdown and audit details.
type ISPUParameterResult struct {
	ID                        uint          `gorm:"primaryKey" json:"id"`
	CalculationResultID       uint          `gorm:"index;not null" json:"calculation_result_id"`
	PollutantCode             ISPUParameter `gorm:"size:16;not null" json:"pollutant_code"`
	ParameterID               uint          `gorm:"index;not null" json:"parameter_id"`
	Parameter                 *Parameter    `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`
	Concentration             float64       `gorm:"not null" json:"concentration"`
	SourceUnit                string        `gorm:"size:32;not null" json:"source_unit"`
	NormalizedConcentration   float64       `gorm:"not null" json:"normalized_concentration"`
	NormalizedUnit            string        `gorm:"size:32;not null;default:'µg/m³'" json:"normalized_unit"`
	CalculatedISPU            float64       `gorm:"not null" json:"calculated_ispu"`
	Category                  ISPUCategory  `gorm:"size:32;not null" json:"category"`
	AggregationConcentration  float64       `gorm:"default:0" json:"aggregation_concentration"`
	AggregationISPU           float64       `gorm:"default:0" json:"aggregation_ispu"`
	AggregationCategory       ISPUCategory  `gorm:"size:32;default:'TIDAK_TERSEDIA'" json:"aggregation_category"`
	Ia                        float64       `json:"ia"`
	Ib                        float64       `json:"ib"`
	Xa                        float64       `json:"xa"`
	Xb                        float64       `json:"xb"`
	CompletenessPercent       float64       `json:"completeness_percent"`
	ValidSamples              int           `json:"valid_samples"`
	TotalExpectedSamples      int           `json:"total_expected_samples"`
	Formula                   string        `gorm:"size:255" json:"formula"`
	DetailsJSON               string        `gorm:"type:text" json:"details_json,omitempty"`
	SourceAggregationCount    int           `gorm:"default:0" json:"source_aggregation_count"`
	SourceLatestIdentifier    string        `gorm:"size:32" json:"source_latest_identifier"`
	SourceAggregationInterval int           `gorm:"default:120" json:"source_aggregation_interval"`
	SourceAggregationFunction string        `gorm:"size:32;default:'AVG'" json:"source_aggregation_function"`
	CreatedAt                 time.Time     `json:"created_at"`
	UpdatedAt                 time.Time     `json:"updated_at"`
}
