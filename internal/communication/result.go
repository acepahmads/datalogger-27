package communication

import (
	"encoding/hex"
	"strings"
	"time"

	"datalogger/internal/model"
)

// CommunicationResult contains the complete telemetry readout metadata and diagnostic info
type CommunicationResult struct {
	DeviceID        uint                   `json:"device_id"`
	ConnectionID    uint                   `json:"connection_id"`
	ParameterID     uint                   `json:"parameter_id"`
	ParameterCode   string                 `json:"parameter_code"`
	Timestamp       time.Time              `json:"timestamp"`
	FunctionCode    byte                   `json:"function_code"`
	RegisterAddress uint16                 `json:"register_address"`
	RegisterCount   uint16                 `json:"register_count"`
	RawBytes        []byte                 `json:"-"`
	RawHex          string                 `json:"raw_hex"`
	RawValue        float64                `json:"raw_value"`
	ScaledValue     float64                `json:"scaled_value"`
	DecodedValue    float64                `json:"decoded_value"`
	Formula         string                 `json:"formula,omitempty"`
	FormulaValue    *float64               `json:"formula_value,omitempty"`
	IsHeld          bool                   `json:"is_held"`
	Quality         model.TelemetryQuality `json:"quality"`
	QualityReason   model.QualityReason    `json:"quality_reason"`
	QualityFlags    string                 `json:"quality_flags,omitempty"`
	ProcessedValue  float64                `json:"processed_value"`
	Success         bool                   `json:"success"`
	ErrorCode       string                 `json:"error_code,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	ResponseTimeMs  int64                  `json:"response_time_ms"`
}

// NewSuccessResult constructs a successful communication result
func NewSuccessResult(
	deviceID, connID, paramID uint,
	paramCode string,
	fc byte,
	addr, count uint16,
	rawBytes []byte,
	rawVal, decodedVal float64,
	responseTime time.Duration,
) *CommunicationResult {
	hexStr := strings.ToUpper(hex.EncodeToString(rawBytes))
	// Add spacing every 2 characters for clean industrial presentation (e.g. "42 F6 CC CD")
	var formattedHex strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 {
			formattedHex.WriteString(" ")
		}
		if i+2 <= len(hexStr) {
			formattedHex.WriteString(hexStr[i : i+2])
		} else {
			formattedHex.WriteString(hexStr[i:])
		}
	}

	return &CommunicationResult{
		DeviceID:        deviceID,
		ConnectionID:    connID,
		ParameterID:     paramID,
		ParameterCode:   paramCode,
		Timestamp:       time.Now(),
		FunctionCode:    fc,
		RegisterAddress: addr,
		RegisterCount:   count,
		RawBytes:        rawBytes,
		RawHex:          formattedHex.String(),
		RawValue:        rawVal,
		DecodedValue:    decodedVal,
		Success:         true,
		ResponseTimeMs:  responseTime.Milliseconds(),
	}
}

// NewErrorResult constructs a failed communication result
func NewErrorResult(
	deviceID, connID, paramID uint,
	paramCode string,
	fc byte,
	addr, count uint16,
	errCode, errMsg string,
	responseTime time.Duration,
) *CommunicationResult {
	return &CommunicationResult{
		DeviceID:        deviceID,
		ConnectionID:    connID,
		ParameterID:     paramID,
		ParameterCode:   paramCode,
		Timestamp:       time.Now(),
		FunctionCode:    fc,
		RegisterAddress: addr,
		RegisterCount:   count,
		Success:         false,
		ErrorCode:       errCode,
		ErrorMessage:    errMsg,
		ResponseTimeMs:  responseTime.Milliseconds(),
	}
}
