package modbus

import (
	"datalogger/internal/model"
)

// BatchMode defines parameter-level participation in multi-address batch read
type BatchMode string

const (
	BatchModeAuto   BatchMode = "AUTO"
	BatchModeSingle BatchMode = "SINGLE"
)

// SlicedParameter represents a parameter mapped into a parent BatchJob
type SlicedParameter struct {
	Parameter     *model.Parameter `json:"parameter"`
	PDUAddress    uint16           `json:"pdu_address"`
	RegisterCount uint16           `json:"register_count"`
	ByteOffset    int              `json:"byte_offset"` // Offset in bytes within the Modbus response Data
	ByteLength    int              `json:"byte_length"` // Byte length to slice for this parameter
	BitOffset     int              `json:"bit_offset"`  // Bit offset for coils (FC01) and discrete inputs (FC02)
}

// BatchJob represents a single consolidated Modbus read request executing on the bus
type BatchJob struct {
	ID              string            `json:"id"`
	DeviceID        uint              `json:"device_id"`
	DeviceCode      string            `json:"device_code"`
	SlaveID         byte              `json:"slave_id"`
	FunctionCode    byte              `json:"function_code"`
	RegisterType    string            `json:"register_type"`
	StartingAddress uint16            `json:"starting_address"` // 0-based PDU address
	Quantity        uint16            `json:"quantity"`         // Number of registers or coils to read
	DisplayAddress  int               `json:"display_address"`  // Modicon convention address for UI display
	DisplayEndAddr  int               `json:"display_end_addr"` // End Modicon address for UI display
	Parameters      []SlicedParameter `json:"parameters"`
	IsSingle        bool              `json:"is_single"`        // True if read individually (opt-out or non-contiguous)
}

// InvalidParameterInfo documents parameters that could not be mapped into the Modbus polling plan
type InvalidParameterInfo struct {
	ParameterID   uint   `json:"parameter_id"`
	ParameterCode string `json:"parameter_code"`
	ParameterName string `json:"parameter_name"`
	Reason        string `json:"reason"`
}

// BatchPlan encapsulates the complete calculated polling plan for a device
type BatchPlan struct {
	DeviceID                  uint                   `json:"device_id"`
	DeviceCode                string                 `json:"device_code"`
	BatchReadEnabled          bool                   `json:"batch_read_enabled"`
	MaxBatchRegisters         int                    `json:"max_batch_registers"`
	MaxRegisterGap            int                    `json:"max_register_gap"`
	TotalParameters           int                    `json:"total_parameters"`
	TotalConfiguredParameters int                    `json:"total_configured_parameters"`
	TotalModbusRequests       int                    `json:"total_modbus_requests"`
	EstimatedReductionPercent float64                `json:"estimated_reduction_percent"`
	Jobs                      []BatchJob             `json:"jobs"`
	InvalidParameters         []InvalidParameterInfo `json:"invalid_parameters,omitempty"`
	EmptyReason               string                 `json:"empty_reason,omitempty"`
}
