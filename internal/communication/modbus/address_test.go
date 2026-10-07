package modbus

import (
	"testing"
)

func TestAddressResolution(t *testing.T) {
	// 1. Modicon 5-digit Holding Register (40001 -> Offset 0, FC 03)
	res1, err := ResolveRegisterAddress("HOLDING_REGISTER", 40001)
	if err != nil || res1.PDUAddress != 0 || res1.FunctionCode != FunctionReadHoldingRegisters {
		t.Fatalf("failed resolving 40001: got addr %d, fc %d, err %v", res1.PDUAddress, res1.FunctionCode, err)
	}

	// 2. Modicon 6-digit Holding Register (400100 -> Offset 99, FC 03)
	res2, err := ResolveRegisterAddress("", 400100)
	if err != nil || res2.PDUAddress != 99 || res2.FunctionCode != FunctionReadHoldingRegisters {
		t.Fatalf("failed resolving 400100: got addr %d, fc %d, err %v", res2.PDUAddress, res2.FunctionCode, err)
	}

	// 3. Modicon 5-digit Input Register (30005 -> Offset 4, FC 04)
	res3, err := ResolveRegisterAddress("", 30005)
	if err != nil || res3.PDUAddress != 4 || res3.FunctionCode != FunctionReadInputRegisters {
		t.Fatalf("failed resolving 30005: got addr %d, fc %d, err %v", res3.PDUAddress, res3.FunctionCode, err)
	}

	// 4. Modicon 5-digit Discrete Input (10010 -> Offset 9, FC 02)
	res4, err := ResolveRegisterAddress("", 10010)
	if err != nil || res4.PDUAddress != 9 || res4.FunctionCode != FunctionReadDiscreteInputs {
		t.Fatalf("failed resolving 10010: got addr %d, fc %d, err %v", res4.PDUAddress, res4.FunctionCode, err)
	}

	// 5. Direct 0-based offset with explicit type (Holding Register, Addr 120 -> Offset 120, FC 03)
	res5, err := ResolveRegisterAddress(RegTypeHoldingRegister, 120)
	if err != nil || res5.PDUAddress != 120 || res5.FunctionCode != FunctionReadHoldingRegisters {
		t.Fatalf("failed resolving direct offset 120: got addr %d, fc %d, err %v", res5.PDUAddress, res5.FunctionCode, err)
	}

	// 6. Direct Coil (Coil 1 -> Offset 0, FC 01)
	res6, err := ResolveRegisterAddress(RegTypeCoil, 1)
	if err != nil || res6.PDUAddress != 0 || res6.FunctionCode != FunctionReadCoils {
		t.Fatalf("failed resolving coil 1: got addr %d, fc %d, err %v", res6.PDUAddress, res6.FunctionCode, err)
	}
}
