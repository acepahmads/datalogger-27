package modbus

import (
	"time"
)

// DiagnosticStageStatus enumerates the outcome of a diagnostic evaluation stage
type DiagnosticStageStatus string

const (
	StagePass                 DiagnosticStageStatus = "PASS"
	StageFail                 DiagnosticStageStatus = "FAIL"
	StageInsufficientEvidence DiagnosticStageStatus = "INSUFFICIENT_EVIDENCE"
	StageNotTested            DiagnosticStageStatus = "NOT_TESTED"
)

// FailureCategory classifies the exact physical or logical root-cause of communication failure
type FailureCategory string

const (
	FailCategoryNone             FailureCategory = "NONE"
	FailCategoryPortNotFound     FailureCategory = "PORT_NOT_FOUND"
	FailCategoryBrokenSymlink    FailureCategory = "BROKEN_SYMLINK"
	FailCategoryPermissionDenied FailureCategory = "PORT_PERMISSION_DENIED"
	FailCategoryPortBusy         FailureCategory = "PORT_BUSY"
	FailCategoryPortOpenFailed   FailureCategory = "PORT_OPEN_FAILED"
	FailCategoryTransmitFailed   FailureCategory = "TRANSMISSION_FAILED"
	FailCategoryTimeout          FailureCategory = "TIMEOUT_NO_RESPONSE"
	FailCategoryInvalidCRC       FailureCategory = "INVALID_CRC"
	FailCategoryModbusException  FailureCategory = "MODBUS_EXCEPTION"
	FailCategoryFrameCorrupted   FailureCategory = "FRAME_CORRUPTED"
	FailCategoryRegisterConfig   FailureCategory = "REGISTER_CONFIG_ERROR"
	FailCategoryDecodeError      FailureCategory = "DECODE_ERROR"
)

// DiagnosticStageEvidence holds traceable results for an individual diagnostic pipeline stage
type DiagnosticStageEvidence struct {
	Stage   string                `json:"stage"`   // PORT_VERIFICATION, SERIAL_CONFIGURATION, MODBUS_TRANSACTION, DATA_INTERPRETATION
	Status  DiagnosticStageStatus `json:"status"`  // PASS, FAIL, INSUFFICIENT_EVIDENCE, NOT_TESTED
	Message string                `json:"message"`
	Details map[string]interface{}`json:"details,omitempty"`
}

// PortEvidence holds physical filesystem and OS port status
type PortEvidence struct {
	ConfiguredPort       string   `json:"configured_port"`
	ResolvedPort         string   `json:"resolved_port"`
	Exists               bool     `json:"exists"`
	IsSymlink            bool     `json:"is_symlink"`
	SymlinkTarget        string   `json:"symlink_target,omitempty"`
	SymlinkTargetExists  bool     `json:"symlink_target_exists"`
	CanAccess            bool     `json:"can_access"`
	PermissionError      string   `json:"permission_error,omitempty"`
	IsBusy               bool     `json:"is_busy"`
	BusyReason           string   `json:"busy_reason,omitempty"`
	AdapterOwned         bool     `json:"adapter_owned"`
	SystemAvailablePorts []string `json:"system_available_ports"`
}

// SerialConfigEvidence captures serial framing and timing
type SerialConfigEvidence struct {
	BaudRate  int    `json:"baud_rate"`
	DataBits  int    `json:"data_bits"`
	Parity    string `json:"parity"`
	StopBits  int    `json:"stop_bits"`
	TimeoutMs int    `json:"timeout_ms"`
}

// ModbusTransactionEvidence captures raw frame bytes and protocol state
type ModbusTransactionEvidence struct {
	SlaveID              byte     `json:"slave_id"`
	FunctionCode         byte     `json:"function_code"`
	RegisterType         string   `json:"register_type"`
	ConfiguredAddress    uint16   `json:"configured_address"`
	PDUAddress           uint16   `json:"pdu_address"`
	Quantity             uint16   `json:"quantity"`
	AddressingConvention string   `json:"addressing_convention"` // 0-BASED_PDU vs 1-BASED_PLC
	RequestTransmitted   bool     `json:"request_transmitted"`
	RequestFrameHex      string   `json:"request_frame_hex,omitempty"`
	RequestFrameBytes    []string `json:"request_frame_bytes,omitempty"`
	ResponseReceived     bool     `json:"response_received"`
	ResponseFrameHex     string   `json:"response_frame_hex,omitempty"`
	ResponseFrameBytes   []string `json:"response_frame_bytes,omitempty"`
	CrcValidation        string   `json:"crc_validation"` // PASS, FAIL, NOT_CHECKED, NOT_APPLICABLE
	CrcExpectedHex       string   `json:"crc_expected_hex,omitempty"`
	CrcReceivedHex       string   `json:"crc_received_hex,omitempty"`
	ModbusExceptionCode  byte     `json:"modbus_exception_code,omitempty"`
	ModbusExceptionName  string   `json:"modbus_exception_name,omitempty"`
}

// DataInterpretationEvidence captures register decoding and engineering scaling
type DataInterpretationEvidence struct {
	RawBytesHex      string   `json:"raw_bytes_hex,omitempty"`
	RawRegisterWords []uint16 `json:"raw_register_words,omitempty"`
	RawRegisterHex   []string `json:"raw_register_hex,omitempty"`
	DataType         string   `json:"data_type"`
	ByteOrder        string   `json:"byte_order"`
	RawValue         *float64 `json:"raw_value,omitempty"`
	DecodedValue     *float64 `json:"decoded_value,omitempty"`
	Scale            float64  `json:"scale"`
	Offset           float64  `json:"offset"`
	ScaledValue      *float64 `json:"scaled_value,omitempty"`
	Formula          string   `json:"formula,omitempty"`
	FormulaValue     *float64 `json:"formula_value,omitempty"`
	EngineeringValue *float64 `json:"engineering_value,omitempty"`
	Unit             string   `json:"unit"`
}

// EngineeringDiagnosticReport provides comprehensive, structured evidence of an on-demand Modbus read
type EngineeringDiagnosticReport struct {
	DiagnosticID         string                      `json:"diagnostic_id"`
	Timestamp            time.Time                   `json:"timestamp"`
	DeviceID             uint                        `json:"device_id"`
	DeviceCode           string                      `json:"device_code"`
	DeviceName           string                      `json:"device_name"`
	ParameterID          uint                        `json:"parameter_id"`
	ParameterCode        string                      `json:"parameter_code"`
	ParameterName        string                      `json:"parameter_name"`
	Protocol             string                      `json:"protocol"`
	Transport            string                      `json:"transport"`
	OverallStatus        string                      `json:"overall_status"` // PASS or FAIL
	Success              bool                        `json:"success"`
	ResponseTimeMs       int64                       `json:"response_time_ms"`
	FailureCategory      FailureCategory             `json:"failure_category"`
	ErrorMessage         string                      `json:"error_message,omitempty"`
	LastSuccessfulStage  string                      `json:"last_successful_stage"`
	FirstFailedStage     string                      `json:"first_failed_stage,omitempty"`
	EvidenceCompleteness string                      `json:"evidence_completeness"` // FULL, PARTIAL, NONE
	SuggestedAction      string                      `json:"suggested_action"`
	SuggestedActionID    string                      `json:"suggested_action_id"`
	MbpollCommand        string                      `json:"mbpoll_command"`
	PortEvidence         PortEvidence                `json:"port_evidence"`
	SerialConfig         SerialConfigEvidence        `json:"serial_config"`
	TransactionEvidence  ModbusTransactionEvidence   `json:"transaction_evidence"`
	DataEvidence         DataInterpretationEvidence  `json:"data_evidence"`
	Stages               []DiagnosticStageEvidence   `json:"stages"`

	// Legacy / Backward Compatibility fields for existing UI components
	RawValue       float64  `json:"raw_value"`
	DecodedValue   float64  `json:"decoded_value"`
	ScaledValue    float64  `json:"scaled_value"`
	Formula        string   `json:"formula,omitempty"`
	FormulaValue   *float64 `json:"formula_value,omitempty"`
	RawBytesHex    string   `json:"raw_bytes_hex,omitempty"`
	FunctionCode   byte     `json:"function_code"`
	RegisterAddress uint16  `json:"register_address"`
	RegisterCount  uint16   `json:"register_count"`
}
