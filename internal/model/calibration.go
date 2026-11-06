package model

import (
	"strings"
	"time"
)

const (
	CalibrationDomainAQMS       = "AQMS"
	CalibrationDomainSparing    = "SPARING"
	CalibrationDomainWQMSOnlimo = "WQMS_ONLIMO"
	CalibrationDomainGeneral    = "GENERAL"

	CalibrationActionApply    = "APPLY"
	CalibrationActionReset    = "RESET"
	CalibrationActionRollback = "ROLLBACK"
)

// SensorCalibrationLog stores an immutable audit record of every calibration, reset, or rollback action
type SensorCalibrationLog struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	DeviceID      uint   `gorm:"index;not null" json:"device_id"`
	DeviceCode    string `gorm:"size:64" json:"device_code"`
	DeviceName    string `gorm:"size:128" json:"device_name"`
	ParameterID   uint   `gorm:"index;not null" json:"parameter_id"`
	ParameterCode string `gorm:"size:64" json:"parameter_code"`
	ParameterName string `gorm:"size:128" json:"parameter_name"`
	Unit          string `gorm:"size:32" json:"unit"`
	Domain        string `gorm:"size:32;index;default:'SPARING'" json:"domain"` // AQMS, SPARING, WQMS_ONLIMO, GENERAL
	Action        string `gorm:"size:32;index;not null" json:"action"`          // APPLY, RESET, ROLLBACK

	CalibrationModel string  `gorm:"size:32;not null" json:"calibration_model"` // OFFSET_1PT, SPAN_1PT, LINEAR, QUADRATIC, NONE
	PointCount       int     `gorm:"default:0" json:"point_count"`
	TotalPoints      int     `gorm:"default:0" json:"total_points"`
	C2               float64 `gorm:"type:double;default:0" json:"c2"`
	C1               float64 `gorm:"type:double;default:1" json:"c1"`
	C0               float64 `gorm:"type:double;default:0" json:"c0"`
	Equation         string  `gorm:"size:160" json:"equation"`

	PrevCalibrationEnabled bool    `gorm:"default:false" json:"prev_calibration_enabled"`
	PrevCalibrationModel   string  `gorm:"size:32" json:"prev_calibration_model"`
	PrevC2                 float64 `gorm:"type:double;default:0" json:"prev_c2"`
	PrevC1                 float64 `gorm:"type:double;default:1" json:"prev_c1"`
	PrevC0                 float64 `gorm:"type:double;default:0" json:"prev_c0"`
	PrevEquation           string  `gorm:"size:160" json:"prev_equation"`
	PrevPointsJSON         string  `gorm:"type:text" json:"prev_points_json,omitempty"`
	PrevStandardNote       string  `gorm:"size:255" json:"prev_standard_note,omitempty"`

	R2           *float64 `gorm:"type:double" json:"r2,omitempty"`
	MinR2        float64  `gorm:"type:double;default:0.995" json:"min_r2"`
	PassedR2     bool     `gorm:"default:true" json:"passed_r2"`
	RMSE         *float64 `gorm:"type:double" json:"rmse,omitempty"`
	PointsJSON   string   `gorm:"type:text" json:"points_json,omitempty"`
	StandardNote string   `gorm:"size:255" json:"standard_note,omitempty"`

	PerformedBy   string    `gorm:"size:64;index" json:"performed_by"`
	PerformedRole string    `gorm:"size:32" json:"performed_role"`
	IPAddress     string    `gorm:"size:64" json:"ip_address,omitempty"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

func (SensorCalibrationLog) TableName() string {
	return "sensor_calibration_logs"
}

func (l *SensorCalibrationLog) CalibrationEnabledFromAction() bool {
	if l == nil {
		return false
	}
	if strings.EqualFold(l.Action, CalibrationActionReset) || strings.EqualFold(l.CalibrationModel, "NONE") {
		return false
	}
	return true
}

// NormalizeCalibrationDomain standardizes domain labels for AQMS, SPARING, and WQMS ONLIMO
func NormalizeCalibrationDomain(d string) string {
	raw := strings.ToUpper(strings.TrimSpace(d))
	switch raw {
	case "AQMS", "ISPU", "UDARA", "AIR_QUALITY":
		return CalibrationDomainAQMS
	case "SPARING", "LIMBAH", "WWTP", "IPAL":
		return CalibrationDomainSparing
	case "WQMS_ONLIMO", "WQMS", "ONLIMO", "WQMS-ONLIMO", "WATER_QUALITY":
		return CalibrationDomainWQMSOnlimo
	case "GENERAL", "OTHER":
		return CalibrationDomainGeneral
	default:
		if strings.Contains(raw, "AQMS") || strings.Contains(raw, "ISPU") {
			return CalibrationDomainAQMS
		}
		if strings.Contains(raw, "ONLIMO") || strings.Contains(raw, "WQMS") {
			return CalibrationDomainWQMSOnlimo
		}
		if strings.Contains(raw, "SPARING") {
			return CalibrationDomainSparing
		}
		return ""
	}
}

// InferCalibrationDomain determines the best compliance domain (AQMS, SPARING, WQMS_ONLIMO) for a parameter
func InferCalibrationDomain(explicitDomain, paramCode, paramName, unit, stationType string) string {
	if norm := NormalizeCalibrationDomain(explicitDomain); norm != "" {
		return norm
	}

	token := strings.ToUpper(paramCode + " " + paramName + " " + unit)
	stNorm := NormalizeCalibrationDomain(stationType)

	// 1. Air Quality / AQMS indicators
	aqmsTokens := []string{
		"PM10", "PM2.5", "PM25", "PM_10", "PM_25", "TSP",
		"SO2", "NO2", "NOX", "O3", "OZON", "CO ", "KARBON MONOKSIDA", "CARBON MONOXIDE",
		"HC", "NMHC", "VOC", "TVOC", "H2S", "UG/M3", "µG/M3", "µG/M³", "PPB",
		"WIND", "ANGIN", "HUMIDITY", "KELEMBABAN", "PRESSURE", "TEKANAN UDARA", "SOLAR", "RAIN", "CURAH HUJAN",
	}
	for _, kw := range aqmsTokens {
		if strings.Contains(token, kw) {
			return CalibrationDomainAQMS
		}
	}
	if strings.TrimSpace(strings.ToUpper(paramCode)) == "CO" {
		return CalibrationDomainAQMS
	}

	// 2. Exclusive WQMS ONLIMO river/surface water multiparameter sonde indicators
	onlimoExclusive := []string{
		"DO", "DISSOLVED OXYGEN", "OKSIGEN TERLARUT",
		"DHL", "CONDUCTIVITY", "KONDUKTIVITAS", "US/CM", "µS/CM", "MS/CM",
		"TDS", "TURB", "KEKERUHAN", "NTU", "ORP", "BOD",
		"NO3", "NITRAT", "NITRATE", "SALIN", "CHLOROPHYLL", "KLOROFIL", "TMA",
	}
	for _, kw := range onlimoExclusive {
		if strings.Contains(token, kw) {
			return CalibrationDomainWQMSOnlimo
		}
	}

	// 3. Shared Water Quality parameters (pH, COD, TSS, NH3-N, Debit/Flow, Water Temp)
	waterShared := []string{"PH", "COD", "TSS", "NH3", "AMMONIA", "AMONIA", "DEBIT", "FLOW", "MG/L"}
	for _, kw := range waterShared {
		if strings.Contains(token, kw) {
			if stNorm == CalibrationDomainWQMSOnlimo {
				return CalibrationDomainWQMSOnlimo
			}
			return CalibrationDomainSparing
		}
	}

	if stNorm != "" {
		return stNorm
	}
	return CalibrationDomainSparing
}
