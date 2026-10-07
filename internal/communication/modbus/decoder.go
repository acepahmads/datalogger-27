package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"datalogger/internal/model"
)

// ByteOrder constants
const (
	ByteOrderABCD = "ABCD" // Big Endian (Standard Modbus)
	ByteOrderCDAB = "CDAB" // Word Swap (Mid-Little Endian)
	ByteOrderBADC = "BADC" // Byte Swap (Mid-Big Endian)
	ByteOrderDCBA = "DCBA" // Little Endian (Full Reverse)
)

// NormalizeByteOrder ensures a valid byte order string
func NormalizeByteOrder(order string) string {
	u := strings.ToUpper(strings.TrimSpace(order))
	switch u {
	case ByteOrderABCD, "BIG_ENDIAN", "BE":
		return ByteOrderABCD
	case ByteOrderCDAB, "WORD_SWAP", "WS":
		return ByteOrderCDAB
	case ByteOrderBADC, "BYTE_SWAP", "BS":
		return ByteOrderBADC
	case ByteOrderDCBA, "LITTLE_ENDIAN", "LE":
		return ByteOrderDCBA
	default:
		return ByteOrderABCD
	}
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

	a, b, c, d := raw[0], raw[1], raw[2], raw[3]
	norm := NormalizeByteOrder(order)

	switch norm {
	case ByteOrderABCD: // Standard Big Endian: High Word first, High Byte first
		return []byte{a, b, c, d}, nil
	case ByteOrderCDAB: // Word Swap: Low Word first, High Byte first
		return []byte{c, d, a, b}, nil
	case ByteOrderBADC: // Byte Swap: High Word first, Low Byte first
		return []byte{b, a, d, c}, nil
	case ByteOrderDCBA: // Little Endian: Low Word first, Low Byte first
		return []byte{d, c, b, a}, nil
	default:
		return []byte{a, b, c, d}, nil
	}
}

// OrderBytes16 rearranges 2 raw bytes from wire order to standard Big-Endian representation
func OrderBytes16(raw []byte, order string) ([]byte, error) {
	if len(raw) < 2 {
		return nil, fmt.Errorf("insufficient bytes for 16-bit decoding: need 2, got %d", len(raw))
	}

	a, b := raw[0], raw[1]
	norm := NormalizeByteOrder(order)

	switch norm {
	case ByteOrderBADC, ByteOrderDCBA: // Low Byte first
		return []byte{b, a}, nil
	default: // High Byte first (ABCD, CDAB)
		return []byte{a, b}, nil
	}
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
func DecodeRegisters(raw []byte, dataType model.ParameterDataType, byteOrder string) (*DecodedValue, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty raw bytes cannot be decoded")
	}

	res := &DecodedValue{}

	switch dataType {
	case model.DataTypeBoolean:
		// If 1 byte (from coil/discrete input)
		if len(raw) == 1 {
			b := (raw[0] & 0x01) != 0
			res.BoolValue = b
			if b {
				res.NumericValue = 1.0
				res.RawValue = 1.0
			} else {
				res.NumericValue = 0.0
				res.RawValue = 0.0
			}
			return res, nil
		}
		// If 2 bytes (from holding/input register word)
		ob, err := OrderBytes16(raw, byteOrder)
		if err != nil {
			return nil, err
		}
		u16 := binary.BigEndian.Uint16(ob)
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
		ob, err := OrderBytes16(raw, byteOrder)
		if err != nil {
			return nil, err
		}
		u16 := binary.BigEndian.Uint16(ob)
		res.NumericValue = float64(u16)
		res.RawValue = float64(u16)
		return res, nil

	case model.DataTypeInt16:
		ob, err := OrderBytes16(raw, byteOrder)
		if err != nil {
			return nil, err
		}
		i16 := int16(binary.BigEndian.Uint16(ob))
		res.NumericValue = float64(i16)
		res.RawValue = float64(i16)
		return res, nil

	case model.DataTypeUInt32:
		ob, err := OrderBytes32(raw, byteOrder)
		if err != nil {
			return nil, err
		}
		u32 := binary.BigEndian.Uint32(ob)
		res.NumericValue = float64(u32)
		res.RawValue = float64(u32)
		return res, nil

	case model.DataTypeInt32:
		ob, err := OrderBytes32(raw, byteOrder)
		if err != nil {
			return nil, err
		}
		i32 := int32(binary.BigEndian.Uint32(ob))
		res.NumericValue = float64(i32)
		res.RawValue = float64(i32)
		return res, nil

	case model.DataTypeFloat32:
		ob, err := OrderBytes32(raw, byteOrder)
		if err != nil {
			return nil, err
		}
		u32 := binary.BigEndian.Uint32(ob)
		f32 := math.Float32frombits(u32)
		// Check for NaN or Inf
		if math.IsNaN(float64(f32)) || math.IsInf(float64(f32), 0) {
			res.NumericValue = 0.0
			res.RawValue = 0.0
		} else {
			res.NumericValue = float64(f32)
			res.RawValue = float64(f32)
		}
		return res, nil

	case model.DataTypeFloat64:
		if len(raw) < 8 {
			return nil, fmt.Errorf("insufficient bytes for FLOAT64: need 8, got %d", len(raw))
		}
		// For 64-bit IEEE 754: handle ABCD vs CDAB/DCBA
		var ordered []byte
		norm := NormalizeByteOrder(byteOrder)
		if norm == ByteOrderCDAB || norm == ByteOrderDCBA {
			// Reverse 16-bit words
			ordered = make([]byte, 8)
			for i := 0; i < 4; i++ {
				srcIdx := (3 - i) * 2
				ordered[i*2] = raw[srcIdx]
				ordered[i*2+1] = raw[srcIdx+1]
			}
		} else {
			ordered = raw[:8]
		}
		u64 := binary.BigEndian.Uint64(ordered)
		f64 := math.Float64frombits(u64)
		if math.IsNaN(f64) || math.IsInf(f64, 0) {
			res.NumericValue = 0.0
			res.RawValue = 0.0
		} else {
			res.NumericValue = f64
			res.RawValue = f64
		}
		return res, nil

	case model.DataTypeString:
		cleanStr := strings.TrimRight(string(raw), "\x00 ")
		res.StringValue = cleanStr
		res.RawValue = float64(len(cleanStr))
		return res, nil

	default:
		// Default fallback to FLOAT32
		ob, err := OrderBytes32(raw, byteOrder)
		if err != nil {
			return nil, err
		}
		u32 := binary.BigEndian.Uint32(ob)
		f32 := math.Float32frombits(u32)
		res.NumericValue = float64(f32)
		res.RawValue = float64(f32)
		return res, nil
	}
}
