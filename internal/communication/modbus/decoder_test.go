package modbus

import (
	"encoding/binary"
	"math"
	"testing"

	"datalogger/internal/model"
)

func TestDecoderTypes(t *testing.T) {
	// 1. UINT16
	rawU16 := []byte{0x12, 0x34}
	decU16, err := DecodeRegisters(rawU16, model.DataTypeUInt16, ByteOrderABCD)
	if err != nil {
		t.Fatalf("UINT16 decode error: %v", err)
	}
	if decU16.NumericValue != 0x1234 {
		t.Fatalf("expected 0x1234 (4660), got %v", decU16.NumericValue)
	}

	// 2. INT16 Negative
	rawI16 := []byte{0xFF, 0xCE} // -50 in two's complement
	decI16, err := DecodeRegisters(rawI16, model.DataTypeInt16, ByteOrderABCD)
	if err != nil {
		t.Fatalf("INT16 decode error: %v", err)
	}
	if decI16.NumericValue != -50 {
		t.Fatalf("expected -50, got %v", decI16.NumericValue)
	}

	// 3. UINT32
	rawU32 := []byte{0x00, 0x01, 0x00, 0x00} // 65536
	decU32, err := DecodeRegisters(rawU32, model.DataTypeUInt32, ByteOrderABCD)
	if err != nil {
		t.Fatalf("UINT32 decode error: %v", err)
	}
	if decU32.NumericValue != 65536 {
		t.Fatalf("expected 65536, got %v", decU32.NumericValue)
	}

	// 4. INT32 Negative
	rawI32 := []byte{0xFF, 0xFF, 0xD8, 0xF0} // -10000
	decI32, err := DecodeRegisters(rawI32, model.DataTypeInt32, ByteOrderABCD)
	if err != nil {
		t.Fatalf("INT32 decode error: %v", err)
	}
	if decI32.NumericValue != -10000 {
		t.Fatalf("expected -10000, got %v", decI32.NumericValue)
	}

	// 5. FLOAT32 (220.5 V)
	// 220.5 in IEEE 754 float32 is 0x435C8000
	valF32 := float32(220.5)
	bitsF32 := math.Float32bits(valF32)
	rawF32 := make([]byte, 4)
	binary.BigEndian.PutUint32(rawF32, bitsF32)

	decF32, err := DecodeRegisters(rawF32, model.DataTypeFloat32, ByteOrderABCD)
	if err != nil {
		t.Fatalf("FLOAT32 decode error: %v", err)
	}
	if math.Abs(decF32.NumericValue-220.5) > 0.001 {
		t.Fatalf("expected 220.5, got %v", decF32.NumericValue)
	}

	// 6. BOOLEAN (Coil 1 byte)
	decBool, err := DecodeRegisters([]byte{0x01}, model.DataTypeBoolean, ByteOrderABCD)
	if err != nil || !decBool.BoolValue {
		t.Fatalf("expected boolean true for coil bit 1")
	}

	decBoolFalse, err := DecodeRegisters([]byte{0x00}, model.DataTypeBoolean, ByteOrderABCD)
	if err != nil || decBoolFalse.BoolValue {
		t.Fatalf("expected boolean false for coil bit 0")
	}
}

func TestDecoderByteOrdering(t *testing.T) {
	// Value 123456789 (0x075BCD15)
	// Wire bytes in Big Endian (ABCD): [0x07, 0x5B, 0xCD, 0x15]
	expected := float64(123456789)

	// 1. ABCD (Big Endian)
	rawABCD := []byte{0x07, 0x5B, 0xCD, 0x15}
	dec1, err := DecodeRegisters(rawABCD, model.DataTypeUInt32, ByteOrderABCD)
	if err != nil || dec1.NumericValue != expected {
		t.Fatalf("ABCD order failed: got %v, err %v", dec1.NumericValue, err)
	}

	// 2. CDAB (Word Swap - Low Word First)
	// Low Word: [0xCD, 0x15], High Word: [0x07, 0x5B]
	rawCDAB := []byte{0xCD, 0x15, 0x07, 0x5B}
	dec2, err := DecodeRegisters(rawCDAB, model.DataTypeUInt32, ByteOrderCDAB)
	if err != nil || dec2.NumericValue != expected {
		t.Fatalf("CDAB word swap failed: got %v, err %v", dec2.NumericValue, err)
	}

	// 3. BADC (Byte Swap)
	rawBADC := []byte{0x5B, 0x07, 0x15, 0xCD}
	dec3, err := DecodeRegisters(rawBADC, model.DataTypeUInt32, ByteOrderBADC)
	if err != nil || dec3.NumericValue != expected {
		t.Fatalf("BADC byte swap failed: got %v, err %v", dec3.NumericValue, err)
	}

	// 4. DCBA (Little Endian)
	rawDCBA := []byte{0x15, 0xCD, 0x5B, 0x07}
	dec4, err := DecodeRegisters(rawDCBA, model.DataTypeUInt32, ByteOrderDCBA)
	if err != nil || dec4.NumericValue != expected {
		t.Fatalf("DCBA little endian failed: got %v, err %v", dec4.NumericValue, err)
	}
}

func TestScaleAndOffset(t *testing.T) {
	// Raw: 2200, Scale: 0.1, Offset: 0.0 -> 220.0
	res1 := ApplyScaleAndOffset(2200, 0.1, 0.0, 1)
	if res1 != 220.0 {
		t.Fatalf("expected 220.0, got %v", res1)
	}

	// Raw: 100, Scale: 1.8, Offset: 32.0 (Celsius to Fahrenheit) -> 212.0
	res2 := ApplyScaleAndOffset(100, 1.8, 32.0, 1)
	if res2 != 212.0 {
		t.Fatalf("expected 212.0, got %v", res2)
	}

	// Precision rounding: Raw: 3.1415926, Scale: 1.0, Offset: 0.0, Precision: 2 -> 3.14
	res3 := ApplyScaleAndOffset(3.1415926, 1.0, 0.0, 2)
	if res3 != 3.14 {
		t.Fatalf("expected 3.14, got %v", res3)
	}
}
