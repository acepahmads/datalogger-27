package modbus

import (
	"fmt"
	"strings"
)

// RegisterType constants
const (
	RegTypeHoldingRegister = "HOLDING_REGISTER"
	RegTypeInputRegister   = "INPUT_REGISTER"
	RegTypeCoil            = "COIL"
	RegTypeDiscreteInput   = "DISCRETE_INPUT"
)

// AddressResolution specifies how an address was interpreted
type AddressResolution struct {
	FunctionCode byte
	PDUAddress   uint16
	Convention   string // "PLC_MODICON_4X", "PLC_MODICON_3X", "PLC_MODICON_1X", "PLC_MODICON_0X", or "DIRECT_OFFSET"
}

/*
ResolveRegisterAddress converts a configured register type and address into:
1. The appropriate Modbus Function Code (FC 01, 02, 03, 04)
2. The exact 0-based PDU protocol address (0..65535)

Addressing Rules & Conventions:
-----------------------------------------------------------------------------------------
1. Modicon 5-digit / 6-digit Convention:
   - 40001..49999 or 400001..465536 -> Holding Register (FC 03), PDU = Address - 40001 (or - 400001)
   - 30001..39999 or 300001..365536 -> Input Register (FC 04), PDU = Address - 30001 (or - 300001)
   - 10001..19999 or 100001..165536 -> Discrete Input (FC 02), PDU = Address - 10001 (or - 100001)
   - 1..9999 (when type is Coil)    -> Coil (FC 01), PDU = Address - 1

2. Direct 0-based Offset Convention:
   - When register_type is explicitly set and address is within standard 0..65535,
     the address is used directly as the 0-based PDU offset.
-----------------------------------------------------------------------------------------
*/
func ResolveRegisterAddress(regType string, rawAddress int) (*AddressResolution, error) {
	if rawAddress < 0 || rawAddress > 655356 {
		return nil, fmt.Errorf("register address %d is out of allowable Modbus range (0..655356)", rawAddress)
	}

	normType := strings.ToUpper(strings.TrimSpace(regType))

	// Check 6-digit Modicon notation
	if rawAddress >= 400001 && rawAddress <= 465536 {
		return &AddressResolution{
			FunctionCode: FunctionReadHoldingRegisters,
			PDUAddress:   uint16(rawAddress - 400001),
			Convention:   "PLC_MODICON_4X_6DIGIT",
		}, nil
	}
	if rawAddress >= 300001 && rawAddress <= 365536 {
		return &AddressResolution{
			FunctionCode: FunctionReadInputRegisters,
			PDUAddress:   uint16(rawAddress - 300001),
			Convention:   "PLC_MODICON_3X_6DIGIT",
		}, nil
	}
	if rawAddress >= 100001 && rawAddress <= 165536 {
		return &AddressResolution{
			FunctionCode: FunctionReadDiscreteInputs,
			PDUAddress:   uint16(rawAddress - 100001),
			Convention:   "PLC_MODICON_1X_6DIGIT",
		}, nil
	}

	// Check 5-digit Modicon notation
	if rawAddress >= 40001 && rawAddress <= 49999 {
		return &AddressResolution{
			FunctionCode: FunctionReadHoldingRegisters,
			PDUAddress:   uint16(rawAddress - 40001),
			Convention:   "PLC_MODICON_4X_5DIGIT",
		}, nil
	}
	if rawAddress >= 30001 && rawAddress <= 39999 {
		return &AddressResolution{
			FunctionCode: FunctionReadInputRegisters,
			PDUAddress:   uint16(rawAddress - 30001),
			Convention:   "PLC_MODICON_3X_5DIGIT",
		}, nil
	}
	if rawAddress >= 10001 && rawAddress <= 19999 {
		return &AddressResolution{
			FunctionCode: FunctionReadDiscreteInputs,
			PDUAddress:   uint16(rawAddress - 10001),
			Convention:   "PLC_MODICON_1X_5DIGIT",
		}, nil
	}

	// If explicit register type is configured
	switch normType {
	case RegTypeHoldingRegister, "4X", "HOLDING":
		if rawAddress > 65535 {
			return nil, fmt.Errorf("holding register address %d exceeds 16-bit PDU address limit (65535)", rawAddress)
		}
		return &AddressResolution{
			FunctionCode: FunctionReadHoldingRegisters,
			PDUAddress:   uint16(rawAddress),
			Convention:   "DIRECT_OFFSET_HOLDING",
		}, nil

	case RegTypeInputRegister, "3X", "INPUT":
		if rawAddress > 65535 {
			return nil, fmt.Errorf("input register address %d exceeds 16-bit PDU address limit (65535)", rawAddress)
		}
		return &AddressResolution{
			FunctionCode: FunctionReadInputRegisters,
			PDUAddress:   uint16(rawAddress),
			Convention:   "DIRECT_OFFSET_INPUT",
		}, nil

	case RegTypeCoil, "0X":
		pdu := rawAddress
		conv := "DIRECT_OFFSET_COIL"
		if rawAddress > 0 && rawAddress <= 9999 {
			pdu = rawAddress - 1
			conv = "PLC_MODICON_0X_1BASED"
		}
		if pdu > 65535 {
			return nil, fmt.Errorf("coil address %d exceeds 16-bit PDU address limit (65535)", rawAddress)
		}
		return &AddressResolution{
			FunctionCode: FunctionReadCoils,
			PDUAddress:   uint16(pdu),
			Convention:   conv,
		}, nil

	case RegTypeDiscreteInput, "1X", "DISCRETE":
		if rawAddress > 65535 {
			return nil, fmt.Errorf("discrete input address %d exceeds 16-bit PDU address limit (65535)", rawAddress)
		}
		return &AddressResolution{
			FunctionCode: FunctionReadDiscreteInputs,
			PDUAddress:   uint16(rawAddress),
			Convention:   "DIRECT_OFFSET_DISCRETE",
		}, nil

	default:
		// Default to holding register direct offset if type is unspecified
		if rawAddress > 65535 {
			return nil, fmt.Errorf("address %d exceeds 16-bit PDU address limit (65535)", rawAddress)
		}
		return &AddressResolution{
			FunctionCode: FunctionReadHoldingRegisters,
			PDUAddress:   uint16(rawAddress),
			Convention:   "DEFAULT_HOLDING_OFFSET",
		}, nil
	}
}
