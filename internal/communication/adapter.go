package communication

import (
	"datalogger/internal/communication/modbus"
)

// Re-export core communication interfaces and constants from modbus subpackage
type ProtocolAdapter = modbus.ProtocolAdapter
type ConnectionState = modbus.ConnectionState
type AdapterStatus = modbus.AdapterStatus
type ModbusReadRequest = modbus.ModbusReadRequest
type ModbusReadResponse = modbus.ModbusReadResponse
type DiagnosticCode = modbus.DiagnosticCode
type DiagnosticResult = modbus.DiagnosticResult
type DiagnosticStageStatus = modbus.DiagnosticStageStatus
type FailureCategory = modbus.FailureCategory
type EngineeringDiagnosticReport = modbus.EngineeringDiagnosticReport
type PortEvidence = modbus.PortEvidence
type SerialConfigEvidence = modbus.SerialConfigEvidence
type ModbusTransactionEvidence = modbus.ModbusTransactionEvidence
type DataInterpretationEvidence = modbus.DataInterpretationEvidence
type DiagnosticStageEvidence = modbus.DiagnosticStageEvidence

const (
	StagePass                 = modbus.StagePass
	StageFail                 = modbus.StageFail
	StageInsufficientEvidence = modbus.StageInsufficientEvidence
	StageNotTested            = modbus.StageNotTested

	FailCategoryNone             = modbus.FailCategoryNone
	FailCategoryPortNotFound     = modbus.FailCategoryPortNotFound
	FailCategoryBrokenSymlink    = modbus.FailCategoryBrokenSymlink
	FailCategoryPermissionDenied = modbus.FailCategoryPermissionDenied
	FailCategoryPortBusy         = modbus.FailCategoryPortBusy
	FailCategoryPortOpenFailed   = modbus.FailCategoryPortOpenFailed
	FailCategoryTransmitFailed   = modbus.FailCategoryTransmitFailed
	FailCategoryTimeout          = modbus.FailCategoryTimeout
	FailCategoryInvalidCRC       = modbus.FailCategoryInvalidCRC
	FailCategoryModbusException  = modbus.FailCategoryModbusException
	FailCategoryFrameCorrupted   = modbus.FailCategoryFrameCorrupted
	FailCategoryRegisterConfig   = modbus.FailCategoryRegisterConfig
	FailCategoryDecodeError      = modbus.FailCategoryDecodeError
)

var InspectSerialPort = modbus.InspectSerialPort

const (
	StateDisconnected ConnectionState = modbus.StateDisconnected
	StateConnecting   ConnectionState = modbus.StateConnecting
	StateConnected    ConnectionState = modbus.StateConnected
	StateDegraded     ConnectionState = modbus.StateDegraded
	StateReconnecting ConnectionState = modbus.StateReconnecting
	StateError        ConnectionState = modbus.StateError
	StateDisabled     ConnectionState = modbus.StateDisabled

	FunctionReadCoils              byte = modbus.FunctionReadCoils
	FunctionReadDiscreteInputs     byte = modbus.FunctionReadDiscreteInputs
	FunctionReadHoldingRegisters   byte = modbus.FunctionReadHoldingRegisters
	FunctionReadInputRegisters     byte = modbus.FunctionReadInputRegisters
	FunctionWriteSingleCoil        byte = modbus.FunctionWriteSingleCoil
	FunctionWriteSingleRegister    byte = modbus.FunctionWriteSingleRegister
	FunctionWriteMultipleCoils     byte = modbus.FunctionWriteMultipleCoils
	FunctionWriteMultipleRegisters byte = modbus.FunctionWriteMultipleRegisters

	DiagSuccess            DiagnosticCode = modbus.DiagSuccess
	DiagTimeout            DiagnosticCode = modbus.DiagTimeout
	DiagSerialOpenError    DiagnosticCode = modbus.DiagSerialOpenError
	DiagSerialIOError      DiagnosticCode = modbus.DiagSerialIOError
	DiagCRCError           DiagnosticCode = modbus.DiagCRCError
	DiagModbusException    DiagnosticCode = modbus.DiagModbusException
	DiagConfigurationError DiagnosticCode = modbus.DiagConfigurationError
	DiagBusy               DiagnosticCode = modbus.DiagBusy
)

// ModbusExceptionName aliases the modbus helper
var ModbusExceptionName = modbus.ModbusExceptionName
