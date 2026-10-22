package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"datalogger/internal/model"
)

// NormalizeByteOrder ensures a valid byte order string for legacy callers
func NormalizeByteOrder(order string) string {
	parsed, err := ParseAndValidateByteOrder(order, model.DataTypeFloat32)
	if err != nil {
		return ByteOrderABCD
	}
	return parsed.Canonical
}

// RequiredRegisterCount returns the number of 16-bit Modbus registers needed for a data type
func RequiredRegisterCount(dataType model.ParameterDataType) uint16 {
	switch dataType {
	case model.DataTypeFloat64:
		return 4 // 8 bytes = 4 registers
	case model.DataTypeFloat32, model.DataTypeInt32, model.DataTypeUInt32:
		return 2 // 4 bytes = 2 registers
	case model.DataTypeInt16, model.DataTypeUInt16, model.DataTypeBoolean:
		return 1 // 2 bytes = 1 register (or 1 bit for digital)
	case model.DataTypeString:
		return 8 // 16 bytes = 8 registers by default
	default:
		return 2
	}
}

// OrderBytes32 rearranges 4 raw bytes from wire order to standard Big-Endian representation
func OrderBytes32(raw []byte, order string) ([]byte, error) {
	if len(raw) < 4 {
		return nil, fmt.Errorf("insufficient bytes for 32-bit decoding: need 4, got %d", len(raw))
	}
	parsed, err := ParseAndValidateByteOrder(order, model.DataTypeFloat32)
	if err != nil {
		return nil, err
	}
	return PermuteBytes(raw, parsed.Permutation)
}

// OrderBytes16 rearranges 2 raw bytes from wire order to standard Big-Endian representation
func OrderBytes16(raw []byte, order string) ([]byte, error) {
	if len(raw) < 2 {
		return nil, fmt.Errorf("insufficient bytes for 16-bit decoding: need 2, got %d", len(raw))
	}
	parsed, err := ParseAndValidateByteOrder(order, model.DataTypeInt16)
	if err != nil {
		return nil, err
	}
	return PermuteBytes(raw, parsed.Permutation)
}

// DecodeCoilBit extracts a boolean state from a Modbus coil or discrete input bit stream
func DecodeCoilBit(data []byte, bitOffset int) (bool, error) {
	byteIdx := bitOffset / 8
	bitInByte := bitOffset % 8
	if byteIdx >= len(data) {
		return false, fmt.Errorf("coil bit index %d out of bounds for %d byte response", bitOffset, len(data))
	}
	val := (data[byteIdx] & (1 << bitInByte)) != 0
	return val, nil
}

// DecodedValue represents the parsed register result
type DecodedValue struct {
	NumericValue float64
	BoolValue    bool
	StringValue  string
	RawValue     float64
}

// DecodeRegisters extracts numeric, boolean, or string data from raw Modbus register bytes
// using the configured byte permutation and endianness-safe binary decoding.
func DecodeRegisters(raw []byte, dataType model.ParameterDataType, byteOrder string) (*DecodedValue, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty raw bytes cannot be decoded")
	}

	// 1. Digital state from single-byte coil/discrete input
	if dataType == model.DataTypeBoolean && len(raw) == 1 {
		b := (raw[0] & 0x01) != 0
		val := 0.0
		if b {
			val = 1.0
		}
		return &DecodedValue{
			NumericValue: val,
			BoolValue:    b,
			RawValue:     val,
		}, nil
	}

	// 2. ASCII String data type
	if dataType == model.DataTypeString {
		cleanStr := strings.TrimRight(string(raw), "\x00 ")
		return &DecodedValue{
			StringValue:  cleanStr,
			RawValue:     float64(len(cleanStr)),
			NumericValue: float64(len(cleanStr)),
		}, nil
	}

	// 3. Parse and validate byte order strictly against data type
	parsedOrder, err := ParseAndValidateByteOrder(byteOrder, dataType)
	if err != nil {
		return nil, fmt.Errorf("byte order validation failed for %s: %w", dataType, err)
	}

	// 4. Apply byte permutation to align raw wire bytes into standard Big-Endian representation
	buf, err := PermuteBytes(raw, parsedOrder.Permutation)
	if err != nil {
		return nil, fmt.Errorf("byte permutation failed: %w", err)
	}

	res := &DecodedValue{}

	// 5. Decode normalized Big-Endian buffer deterministically across CPU architectures
	switch dataType {
	case model.DataTypeBoolean:
		// 16-bit register word boolean
		u16 := binary.BigEndian.Uint16(buf)
		b := (u16 != 0)
		res.BoolValue = b
		if b {
			res.NumericValue = 1.0
			res.RawValue = float64(u16)
		} else {
			res.NumericValue = 0.0
			res.RawValue = 0.0
		}
		return res, nil

	case model.DataTypeUInt16:
		u16 := binary.BigEndian.Uint16(buf)
		res.NumericValue = float64(u16)
		res.RawValue = float64(u16)
		return res, nil

	case model.DataTypeInt16:
		i16 := int16(binary.BigEndian.Uint16(buf))
		res.NumericValue = float64(i16)
		res.RawValue = float64(i16)
		return res, nil

	case model.DataTypeUInt32:
		u32 := binary.BigEndian.Uint32(buf)
		res.NumericValue = float64(u32)
		res.RawValue = float64(u32)
		return res, nil

	case model.DataTypeInt32:
		i32 := int32(binary.BigEndian.Uint32(buf))
		res.NumericValue = float64(i32)
		res.RawValue = float64(i32)
		return res, nil

	case model.DataTypeFloat32:
		u32 := binary.BigEndian.Uint32(buf)
		f32 := math.Float32frombits(u32)
		if math.IsNaN(float64(f32)) || math.IsInf(float64(f32), 0) {
			res.NumericValue = 0.0
			res.RawValue = 0.0
		} else {
			res.NumericValue = float64(f32)
			res.RawValue = float64(f32)
		}
		return res, nil

	case model.DataTypeFloat64:
		u64 := binary.BigEndian.Uint64(buf)
		f64 := math.Float64frombits(u64)
		if math.IsNaN(f64) || math.IsInf(f64, 0) {
			res.NumericValue = 0.0
			res.RawValue = 0.0
		} else {
			res.NumericValue = f64
			res.RawValue = f64
		}
		return res, nil

	default:
		// Default fallback to FLOAT32 if 4 bytes
		if len(buf) >= 4 {
			u32 := binary.BigEndian.Uint32(buf)
			f32 := math.Float32frombits(u32)
			res.NumericValue = float64(f32)
			res.RawValue = float64(f32)
			return res, nil
		}
		return nil, fmt.Errorf("unsupported parameter data type: %s", dataType)
	}
}
