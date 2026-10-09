package modbus

import (
	"context"
	"fmt"
	"time"

	"datalogger/internal/model"
)

// ConnectionState represents the lifecycle state of a protocol connection
type ConnectionState string

const (
	StateDisconnected ConnectionState = "DISCONNECTED"
	StateConnecting   ConnectionState = "CONNECTING"
	StateConnected    ConnectionState = "CONNECTED"
	StateDegraded     ConnectionState = "DEGRADED"
	StateReconnecting ConnectionState = "RECONNECTING"
	StateError        ConnectionState = "ERROR"
	StateDisabled     ConnectionState = "DISABLED"
)

// Modbus Function Codes
const (
	FunctionReadCoils              byte = 0x01
	FunctionReadDiscreteInputs     byte = 0x02
	FunctionReadHoldingRegisters   byte = 0x03
	FunctionReadInputRegisters     byte = 0x04
	FunctionWriteSingleCoil        byte = 0x05
	FunctionWriteSingleRegister    byte = 0x06
	FunctionWriteMultipleCoils     byte = 0x0F
	FunctionWriteMultipleRegisters byte = 0x10
)

// Modbus Exception Codes
const (
	ExIllegalFunction        byte = 0x01
	ExIllegalDataAddress     byte = 0x02
	ExIllegalDataValue       byte = 0x03
	ExSlaveDeviceFailure     byte = 0x04
	ExAcknowledge            byte = 0x05
	ExSlaveDeviceBusy        byte = 0x06
	ExNegativeAcknowledge    byte = 0x07
	ExMemoryParityError      byte = 0x08
	ExGatewayPathUnavailable byte = 0x0A
	ExGatewayTargetFailed    byte = 0x0B
)

// ModbusExceptionName returns the human-readable description of a Modbus exception
func ModbusExceptionName(code byte) string {
	switch code {
	case ExIllegalFunction:
		return "Illegal Function (01)"
	case ExIllegalDataAddress:
		return "Illegal Data Address (02)"
	case ExIllegalDataValue:
		return "Illegal Data Value (03)"
	case ExSlaveDeviceFailure:
		return "Slave Device Failure (04)"
	case ExAcknowledge:
		return "Acknowledge (05)"
	case ExSlaveDeviceBusy:
		return "Slave Device Busy (06)"
	case ExNegativeAcknowledge:
		return "Negative Acknowledge (07)"
	case ExMemoryParityError:
		return "Memory Parity Error (08)"
	case ExGatewayPathUnavailable:
		return "Gateway Path Unavailable (0A)"
	case ExGatewayTargetFailed:
		return "Gateway Target Device Failed to Respond (0B)"
	default:
		return fmt.Sprintf("Unknown Exception (%02X)", code)
	}
}

// ModbusReadRequest specifies a Modbus read operation
type ModbusReadRequest struct {
	SlaveID         byte
	FunctionCode    byte
	StartingAddress uint16
	Quantity        uint16
}

// ModbusReadResponse contains the raw response payload, latency, and captured frame traces
type ModbusReadResponse struct {
	SlaveID          byte
	FunctionCode     byte
	ByteCount        byte
	Data             []byte
	ResponseTime     time.Duration
	RequestADU       []byte
	ResponseADU      []byte
	CRCExpected      uint16
	CRCReceived      uint16
	CRCPassed        bool
	ExceptionCode    byte
	ExceptionMessage string
}

// ModbusTransactionError carries structured frame information when a Modbus transaction fails
type ModbusTransactionError struct {
	Category         FailureCategory
	Message          string
	RequestADU       []byte
	ResponseADU      []byte
	RequestSent      bool
	ResponseReceived bool
	CRCExpected      uint16
	CRCReceived      uint16
	CRCPassed        bool
	ExceptionCode    byte
	ExceptionMessage string
	Err              error
}

func (e *ModbusTransactionError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "modbus transaction error"
}

func (e *ModbusTransactionError) Unwrap() error {
	return e.Err
}

// AdapterStatus captures the real-time operational state and metrics of a protocol adapter
type AdapterStatus struct {
	State             ConnectionState `json:"state"`
	ConnectedSince    *time.Time      `json:"connected_since,omitempty"`
	LastCommunication *time.Time      `json:"last_communication,omitempty"`
	LatencyMs         int             `json:"latency_ms"`
	RetryCount        int             `json:"retry_count"`
	LastError         string          `json:"last_error,omitempty"`
	SuccessCount      int64           `json:"success_count"`
	FailedCount       int64           `json:"failed_count"`
}

// ProtocolAdapter is the modular abstraction for all industrial edge communication protocols
type ProtocolAdapter interface {
	// Connect establishes the physical/logical connection
	Connect(ctx context.Context) error

	// Disconnect closes the active connection gracefully
	Disconnect() error

	// IsConnected returns whether the connection is active and ready
	IsConnected() bool

	// ReadRegisters executes a Modbus read function (FC 01, 02, 03, 04)
	ReadRegisters(ctx context.Context, req ModbusReadRequest) (*ModbusReadResponse, error)

	// HealthCheck tests if the transport is healthy and responsive
	HealthCheck(ctx context.Context) error

	// GetStatus returns the current status and metrics
	GetStatus() AdapterStatus

	// GetConfig returns the connection configuration
	GetConfig() *model.DeviceConnection
}

// DiagnosticCode enumerates standardized communication diagnostic outcomes
type DiagnosticCode string

const (
	DiagSuccess            DiagnosticCode = "SUCCESS"
	DiagTimeout            DiagnosticCode = "TIMEOUT"
	DiagSerialOpenError    DiagnosticCode = "SERIAL_OPEN_ERROR"
	DiagSerialIOError      DiagnosticCode = "SERIAL_IO_ERROR"
	DiagCRCError           DiagnosticCode = "CRC_ERROR"
	DiagModbusException    DiagnosticCode = "MODBUS_EXCEPTION"
	DiagConfigurationError DiagnosticCode = "CONFIGURATION_ERROR"
	DiagBusy               DiagnosticCode = "BUSY"
)

// DiagnosticResult contains structured diagnostic information returned by Test Link
type DiagnosticResult struct {
	Code            DiagnosticCode `json:"code"`
	Success         bool           `json:"success"`
	Connected       bool           `json:"connected"`
	LatencyMs       int            `json:"latency_ms"`
	Message         string         `json:"message"`
	Error           string         `json:"error,omitempty"`
	SlaveID         byte           `json:"slave_id"`
	FunctionCode    byte           `json:"function_code"`
	StartingAddress uint16         `json:"starting_address"`
	RegisterCount   uint16         `json:"register_count"`
	Port            string         `json:"port,omitempty"`
	BaudRate        int            `json:"baud_rate,omitempty"`
	Timestamp       time.Time      `json:"timestamp"`
}
