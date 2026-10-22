package modbus

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"datalogger/internal/model"
)

func TestDecoderTypes(t *testing.T) {
	// 1. UINT16
	rawU16 := []byte{0x12, 0x34}
	decU16, err := DecodeRegisters(rawU16, model.DataTypeUInt16, ByteOrderAB)
	if err != nil {
		t.Fatalf("UINT16 decode error: %v", err)
	}
	if decU16.NumericValue != 0x1234 {
		t.Fatalf("expected 0x1234 (4660), got %v", decU16.NumericValue)
	}

	// 2. INT16 Negative
	rawI16 := []byte{0xFF, 0xCE} // -50 in two's complement
	decI16, err := DecodeRegisters(rawI16, model.DataTypeInt16, ByteOrderAB)
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

	// 5. FLOAT32 (220.5 V) -> 0x435C8000
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
	decBool, err := DecodeRegisters([]byte{0x01}, model.DataTypeBoolean, ByteOrderAB)
	if err != nil || !decBool.BoolValue {
		t.Fatalf("expected boolean true for coil bit 1")
	}

	decBoolFalse, err := DecodeRegisters([]byte{0x00}, model.DataTypeBoolean, ByteOrderAB)
	if err != nil || decBoolFalse.BoolValue {
		t.Fatalf("expected boolean false for coil bit 0")
	}

	// 7. BOOLEAN (Register word 2 bytes)
	decBoolReg, err := DecodeRegisters([]byte{0x00, 0x01}, model.DataTypeBoolean, ByteOrderAB)
	if err != nil || !decBoolReg.BoolValue || decBoolReg.NumericValue != 1.0 {
		t.Fatalf("expected boolean true for register word, got %v", decBoolReg)
	}
}

func TestDecoderByteOrdering32Bit(t *testing.T) {
	// Value 123456789 (0x075BCD15)
	expected := float64(123456789)

	// 1. ABCD (Big Endian)
	rawABCD := []byte{0x07, 0x5B, 0xCD, 0x15}
	dec1, err := DecodeRegisters(rawABCD, model.DataTypeUInt32, ByteOrderABCD)
	if err != nil || dec1.NumericValue != expected {
		t.Fatalf("ABCD order failed: got %v, err %v", dec1.NumericValue, err)
	}

	// 2. CDAB (Word Swap - Low Word First)
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

func TestDecoderByteOrdering16Bit(t *testing.T) {
	// Value 4660 (0x1234)
	expected := float64(4660)

	// 1. AB (High Byte First)
	rawAB := []byte{0x12, 0x34}
	dec1, err := DecodeRegisters(rawAB, model.DataTypeUInt16, ByteOrderAB)
	if err != nil || dec1.NumericValue != expected {
		t.Fatalf("AB order failed: got %v, err %v", dec1.NumericValue, err)
	}

	// 2. BA (Low Byte First)
	rawBA := []byte{0x34, 0x12}
	dec2, err := DecodeRegisters(rawBA, model.DataTypeUInt16, ByteOrderBA)
	if err != nil || dec2.NumericValue != expected {
		t.Fatalf("BA order failed: got %v, err %v", dec2.NumericValue, err)
	}

	// 3. Negative INT16: -1234 (0xFB2E)
	expNeg := float64(-1234)
	rawNegAB := []byte{0xFB, 0x2E}
	decNeg1, err := DecodeRegisters(rawNegAB, model.DataTypeInt16, ByteOrderAB)
	if err != nil || decNeg1.NumericValue != expNeg {
		t.Fatalf("INT16 AB order failed: got %v, err %v", decNeg1.NumericValue, err)
	}

	rawNegBA := []byte{0x2E, 0xFB}
	decNeg2, err := DecodeRegisters(rawNegBA, model.DataTypeInt16, ByteOrderBA)
	if err != nil || decNeg2.NumericValue != expNeg {
		t.Fatalf("INT16 BA order failed: got %v, err %v", decNeg2.NumericValue, err)
	}
}

func TestDecoderByteOrdering64Bit(t *testing.T) {
	// Value 1234567890123456789 (0x1122334455667715)
	// Wire bytes in Big Endian: [0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x15]
	expected := 123456.7890123 // IEEE 754 Float64
	bitsF64 := math.Float64bits(expected)
	beBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(beBytes, bitsF64)
	a, b, c, d, e, f, g, h := beBytes[0], beBytes[1], beBytes[2], beBytes[3], beBytes[4], beBytes[5], beBytes[6], beBytes[7]

	// 1. ABCDEFGH (Standard Big Endian)
	rawBE := []byte{a, b, c, d, e, f, g, h}
	dec1, err := DecodeRegisters(rawBE, model.DataTypeFloat64, ByteOrderABCDEFGH)
	if err != nil || math.Abs(dec1.NumericValue-expected) > 1e-9 {
		t.Fatalf("ABCDEFGH failed: got %v, err %v", dec1.NumericValue, err)
	}

	// 2. GHEFCDAB (Word Swap - 16-bit word reverse)
	rawWordSwap := []byte{g, h, e, f, c, d, a, b}
	dec2, err := DecodeRegisters(rawWordSwap, model.DataTypeFloat64, ByteOrderGHEFCDAB)
	if err != nil || math.Abs(dec2.NumericValue-expected) > 1e-9 {
		t.Fatalf("GHEFCDAB word swap failed: got %v, err %v", dec2.NumericValue, err)
	}

	// 3. BADCFEHG (Byte Swap within each word)
	rawByteSwap := []byte{b, a, d, c, f, e, h, g}
	dec3, err := DecodeRegisters(rawByteSwap, model.DataTypeFloat64, ByteOrderBADCFEHG)
	if err != nil || math.Abs(dec3.NumericValue-expected) > 1e-9 {
		t.Fatalf("BADCFEHG byte swap failed: got %v, err %v", dec3.NumericValue, err)
	}

	// 4. HGFEDCBA (Little Endian - Full Reverse)
	rawLE := []byte{h, g, f, e, d, c, b, a}
	dec4, err := DecodeRegisters(rawLE, model.DataTypeFloat64, ByteOrderHGFEDCBA)
	if err != nil || math.Abs(dec4.NumericValue-expected) > 1e-9 {
		t.Fatalf("HGFEDCBA little endian failed: got %v, err %v", dec4.NumericValue, err)
	}

	// 5. EFGHABCD (Double Word Swap - 32-bit swap)
	rawDWordSwap := []byte{e, f, g, h, a, b, c, d}
	dec5, err := DecodeRegisters(rawDWordSwap, model.DataTypeFloat64, ByteOrderEFGHABCD)
	if err != nil || math.Abs(dec5.NumericValue-expected) > 1e-9 {
		t.Fatalf("EFGHABCD dword swap failed: got %v, err %v", dec5.NumericValue, err)
	}
}

func TestCustomByteOrdering32Bit(t *testing.T) {
	// Value: 123.456 (FLOAT32)
	// IEEE 754 Big Endian: [0x42, 0xF6, 0xE9, 0x79]
	expected := float64(float32(123.456))
	be := []byte{0x42, 0xF6, 0xE9, 0x79}
	a, b, c, d := be[0], be[1], be[2], be[3]

	// Custom 1: Output CDAB from input raw CDAB
	// Wire has [C, D, A, B] -> custom permutation "CDAB" maps output[0]=raw[2]=A, etc.
	// As defined in spec:
	// Input byte: A B C D
	// Custom output: C D A B
	// Wire has [C, D, A, B]. Permutation "CDAB" produces [A, B, C, D]
	rawWire1 := []byte{c, d, a, b}
	dec1, err := DecodeRegisters(rawWire1, model.DataTypeFloat32, "CDAB")
	if err != nil || math.Abs(dec1.NumericValue-expected) > 1e-4 {
		t.Fatalf("Custom CDAB failed: got %v, err %v", dec1.NumericValue, err)
	}

	// Custom 2: With "CUSTOM:" prefix
	dec2, err := DecodeRegisters(rawWire1, model.DataTypeFloat32, "CUSTOM:CDAB")
	if err != nil || math.Abs(dec2.NumericValue-expected) > 1e-4 {
		t.Fatalf("CUSTOM:CDAB failed: got %v, err %v", dec2.NumericValue, err)
	}

	// Custom 3: Arbitrary custom permutation "BDAC"
	// Output order BDAC:
	// output[0] = raw[1] -> needs to be a
	// output[1] = raw[3] -> needs to be b
	// output[2] = raw[0] -> needs to be c
	// output[3] = raw[2] -> needs to be d
	// So wire raw must be: raw[0]=c, raw[1]=a, raw[2]=d, raw[3]=b -> [c, a, d, b]
	rawWire3 := []byte{c, a, d, b}
	dec3, err := DecodeRegisters(rawWire3, model.DataTypeFloat32, "CUSTOM:BDAC")
	if err != nil || math.Abs(dec3.NumericValue-expected) > 1e-4 {
		t.Fatalf("CUSTOM:BDAC failed: got %v, err %v", dec3.NumericValue, err)
	}

	// Custom 4: Comma-separated spaces "C, D, A, B"
	dec4, err := DecodeRegisters(rawWire1, model.DataTypeFloat32, "CUSTOM: C, D, A, B ")
	if err != nil || math.Abs(dec4.NumericValue-expected) > 1e-4 {
		t.Fatalf("CUSTOM with formatting failed: got %v, err %v", dec4.NumericValue, err)
	}
}

func TestCustomByteOrdering64Bit(t *testing.T) {
	// Value: 987654.3210987 (FLOAT64)
	expected := 987654.3210987
	bitsF64 := math.Float64bits(expected)
	be := make([]byte, 8)
	binary.BigEndian.PutUint64(be, bitsF64)
	a, b, c, d, e, f, g, h := be[0], be[1], be[2], be[3], be[4], be[5], be[6], be[7]

	// Custom 64-bit: "EFGHABCD"
	// Output[0..7] = raw[4], raw[5], raw[6], raw[7], raw[0], raw[1], raw[2], raw[3]
	// If raw wire is [e, f, g, h, a, b, c, d], permutation EFGHABCD restores [a, b, c, d, e, f, g, h]
	rawWire := []byte{e, f, g, h, a, b, c, d}
	dec1, err := DecodeRegisters(rawWire, model.DataTypeFloat64, "CUSTOM:EFGHABCD")
	if err != nil || math.Abs(dec1.NumericValue-expected) > 1e-9 {
		t.Fatalf("CUSTOM:EFGHABCD failed: got %v, err %v", dec1.NumericValue, err)
	}

	// Custom 64-bit with spaces and commas: "E, F, G, H, A, B, C, D"
	dec2, err := DecodeRegisters(rawWire, model.DataTypeFloat64, "CUSTOM: E, F, G, H, A, B, C, D")
	if err != nil || math.Abs(dec2.NumericValue-expected) > 1e-9 {
		t.Fatalf("CUSTOM: E, F, G, H, A, B, C, D failed: got %v, err %v", dec2.NumericValue, err)
	}
}

func TestDecoderIEEE754Vectors(t *testing.T) {
	// Known IEEE 754 test vectors:
	testCases := []struct {
		name      string
		dataType  model.ParameterDataType
		rawBE     []byte
		byteOrder string
		expected  float64
	}{
		{
			name:      "Float32 Zero",
			dataType:  model.DataTypeFloat32,
			rawBE:     []byte{0x00, 0x00, 0x00, 0x00},
			byteOrder: "ABCD",
			expected:  0.0,
		},
		{
			name:      "Float32 Positive 1.0",
			dataType:  model.DataTypeFloat32,
			rawBE:     []byte{0x3F, 0x80, 0x00, 0x00},
			byteOrder: "ABCD",
			expected:  1.0,
		},
		{
			name:      "Float32 Negative -1.0",
			dataType:  model.DataTypeFloat32,
			rawBE:     []byte{0xBF, 0x80, 0x00, 0x00},
			byteOrder: "ABCD",
			expected:  -1.0,
		},
		{
			name:      "Float32 Standard 123.456",
			dataType:  model.DataTypeFloat32,
			rawBE:     []byte{0x42, 0xF6, 0xE9, 0x79},
			byteOrder: "ABCD",
			expected:  float64(float32(123.456)),
		},
		{
			name:      "Float32 Negative -123.456",
			dataType:  model.DataTypeFloat32,
			rawBE:     []byte{0xC2, 0xF6, 0xE9, 0x79},
			byteOrder: "ABCD",
			expected:  float64(float32(-123.456)),
		},
		{
			name:      "Float64 Zero",
			dataType:  model.DataTypeFloat64,
			rawBE:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			byteOrder: "ABCDEFGH",
			expected:  0.0,
		},
		{
			name:      "Float64 Positive 1.0",
			dataType:  model.DataTypeFloat64,
			rawBE:     []byte{0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			byteOrder: "ABCDEFGH",
			expected:  1.0,
		},
		{
			name:      "Float64 Negative -1.0",
			dataType:  model.DataTypeFloat64,
			rawBE:     []byte{0xBF, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			byteOrder: "ABCDEFGH",
			expected:  -1.0,
		},
		{
			name:      "Float64 Pi (3.141592653589793)",
			dataType:  model.DataTypeFloat64,
			rawBE:     []byte{0x40, 0x09, 0x21, 0xFB, 0x54, 0x44, 0x2D, 0x18},
			byteOrder: "ABCDEFGH",
			expected:  math.Pi,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dec, err := DecodeRegisters(tc.rawBE, tc.dataType, tc.byteOrder)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(dec.NumericValue-tc.expected) > 1e-7 {
				t.Fatalf("expected %v, got %v", tc.expected, dec.NumericValue)
			}
		})
	}
}

func TestIntegerRepresentationsAndBoundaries(t *testing.T) {
	// INT16 Min (-32768) and Max (32767)
	rawMinI16 := []byte{0x80, 0x00}
	decMinI16, err := DecodeRegisters(rawMinI16, model.DataTypeInt16, ByteOrderAB)
	if err != nil || decMinI16.NumericValue != -32768 {
		t.Fatalf("INT16 Min failed: got %v, err %v", decMinI16.NumericValue, err)
	}

	rawMaxI16 := []byte{0x7F, 0xFF}
	decMaxI16, err := DecodeRegisters(rawMaxI16, model.DataTypeInt16, ByteOrderAB)
	if err != nil || decMaxI16.NumericValue != 32767 {
		t.Fatalf("INT16 Max failed: got %v, err %v", decMaxI16.NumericValue, err)
	}

	// UINT16 Max (65535)
	rawMaxU16 := []byte{0xFF, 0xFF}
	decMaxU16, err := DecodeRegisters(rawMaxU16, model.DataTypeUInt16, ByteOrderAB)
	if err != nil || decMaxU16.NumericValue != 65535 {
		t.Fatalf("UINT16 Max failed: got %v, err %v", decMaxU16.NumericValue, err)
	}

	// INT32 Min (-2147483648) and Max (2147483647)
	rawMinI32 := []byte{0x80, 0x00, 0x00, 0x00}
	decMinI32, err := DecodeRegisters(rawMinI32, model.DataTypeInt32, ByteOrderABCD)
	if err != nil || decMinI32.NumericValue != -2147483648 {
		t.Fatalf("INT32 Min failed: got %v, err %v", decMinI32.NumericValue, err)
	}

	rawMaxI32 := []byte{0x7F, 0xFF, 0xFF, 0xFF}
	decMaxI32, err := DecodeRegisters(rawMaxI32, model.DataTypeInt32, ByteOrderABCD)
	if err != nil || decMaxI32.NumericValue != 2147483647 {
		t.Fatalf("INT32 Max failed: got %v, err %v", decMaxI32.NumericValue, err)
	}

	// UINT32 Max (4294967295)
	rawMaxU32 := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	decMaxU32, err := DecodeRegisters(rawMaxU32, model.DataTypeUInt32, ByteOrderABCD)
	if err != nil || decMaxU32.NumericValue != 4294967295 {
		t.Fatalf("UINT32 Max failed: got %v, err %v", decMaxU32.NumericValue, err)
	}
}

func TestInvalidByteOrderPermutations(t *testing.T) {
	raw4 := []byte{0x01, 0x02, 0x03, 0x04}
	raw8 := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// 1. Duplicate byte position
	_, err := DecodeRegisters(raw4, model.DataTypeFloat32, "ABCA")
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate byte error for ABCA, got: %v", err)
	}

	_, err = DecodeRegisters(raw8, model.DataTypeFloat64, "AABCDEFG")
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate byte error for AABCDEFG, got: %v", err)
	}

	// 2. Missing position / invalid char outside range
	_, err = DecodeRegisters(raw4, model.DataTypeFloat32, "ABCE")
	if err == nil || !strings.Contains(err.Error(), "invalid byte position") {
		t.Fatalf("expected invalid byte position error for ABCE, got: %v", err)
	}

	// 3. Wrong length: 3 positions for 32-bit (requires 4)
	_, err = DecodeRegisters(raw4, model.DataTypeFloat32, "ABC")
	if err == nil || !strings.Contains(err.Error(), "requires exactly 4") {
		t.Fatalf("expected wrong length error for ABC, got: %v", err)
	}

	// 4. Wrong length: 5 positions for 32-bit
	_, err = DecodeRegisters(raw4, model.DataTypeFloat32, "ABCDE")
	if err == nil || !strings.Contains(err.Error(), "requires exactly 4") {
		t.Fatalf("expected wrong length error for ABCDE, got: %v", err)
	}

	// 5. Completely unrecognized random string
	_, err = DecodeRegisters(raw4, model.DataTypeFloat32, "XYZW")
	if err == nil || !strings.Contains(err.Error(), "invalid byte position") {
		t.Fatalf("expected error for XYZW, got: %v", err)
	}
}

func TestEndianDeterminismAcrossPlatforms(t *testing.T) {
	// Verify that binary.BigEndian is deterministic on any platform (x86_64, aarch64, etc.)
	raw := []byte{0x40, 0x49, 0x0F, 0xDB} // float32 for 3.1415927
	u32 := binary.BigEndian.Uint32(raw)
	f32 := math.Float32frombits(u32)

	dec, err := DecodeRegisters(raw, model.DataTypeFloat32, ByteOrderABCD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dec.NumericValue != float64(f32) {
		t.Fatalf("expected bitwise exact float32 %v, got %v", f32, dec.NumericValue)
	}
}

func TestLegacyCompatibility(t *testing.T) {
	// 1. Empty order string defaults to Big Endian for data type
	raw4 := []byte{0x00, 0x01, 0x00, 0x00}
	decDef32, err := DecodeRegisters(raw4, model.DataTypeUInt32, "")
	if err != nil || decDef32.NumericValue != 65536 {
		t.Fatalf("default empty order for 32-bit failed: got %v, err %v", decDef32, err)
	}

	raw2 := []byte{0x12, 0x34}
	decDef16, err := DecodeRegisters(raw2, model.DataTypeUInt16, "")
	if err != nil || decDef16.NumericValue != 0x1234 {
		t.Fatalf("default empty order for 16-bit failed: got %v, err %v", decDef16, err)
	}

	// 2. Legacy alias strings
	decAliasBE, err := DecodeRegisters(raw4, model.DataTypeUInt32, "BIG_ENDIAN")
	if err != nil || decAliasBE.NumericValue != 65536 {
		t.Fatalf("legacy BIG_ENDIAN failed: got %v, err %v", decAliasBE, err)
	}

	decAliasWS, err := DecodeRegisters([]byte{0x00, 0x00, 0x00, 0x01}, model.DataTypeUInt32, "WORD_SWAP")
	if err != nil || decAliasWS.NumericValue != 65536 {
		t.Fatalf("legacy WORD_SWAP failed: got %v, err %v", decAliasWS, err)
	}

	// 3. Legacy 16-bit parameter stored with 4-letter ABCD from previous system defaults
	dec16Legacy, err := DecodeRegisters(raw2, model.DataTypeUInt16, "ABCD")
	if err != nil || dec16Legacy.NumericValue != 0x1234 {
		t.Fatalf("legacy 16-bit ABCD auto-adaptation failed: got %v, err %v", dec16Legacy, err)
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

	// Negative scale and offset
	res4 := ApplyScaleAndOffset(50.0, -2.0, -10.0, 1)
	if res4 != -110.0 {
		t.Fatalf("expected -110.0, got %v", res4)
	}
}
