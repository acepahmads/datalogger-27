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

const (
	StateDisconnected ConnectionState = modbus.StateDisconnected
	StateConnecting   ConnectionState = modbus.StateConnecting
	StateConnected    ConnectionState = modbus.StateConnected
	StateReconnecting ConnectionState = modbus.StateReconnecting
	StateError        ConnectionState = modbus.StateError

	FunctionReadCoils              byte = modbus.FunctionReadCoils
	FunctionReadDiscreteInputs     byte = modbus.FunctionReadDiscreteInputs
	FunctionReadHoldingRegisters   byte = modbus.FunctionReadHoldingRegisters
	FunctionReadInputRegisters     byte = modbus.FunctionReadInputRegisters
	FunctionWriteSingleCoil        byte = modbus.FunctionWriteSingleCoil
	FunctionWriteSingleRegister    byte = modbus.FunctionWriteSingleRegister
	FunctionWriteMultipleCoils     byte = modbus.FunctionWriteMultipleCoils
	FunctionWriteMultipleRegisters byte = modbus.FunctionWriteMultipleRegisters
)

// ModbusExceptionName aliases the modbus helper
var ModbusExceptionName = modbus.ModbusExceptionName
