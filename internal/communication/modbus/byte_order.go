package modbus

import (
	"fmt"
	"strings"

	"datalogger/internal/model"
)

// Standard ByteOrder preset constants
const (
	// 16-bit presets
	ByteOrderAB = "AB" // Big Endian (High Byte First)
	ByteOrderBA = "BA" // Little Endian (Low Byte First / Byte Swap)

	// 32-bit presets
	ByteOrderABCD = "ABCD" // Big Endian (Standard Modbus / High Word First, High Byte First)
	ByteOrderCDAB = "CDAB" // Word Swap (Mid-Little Endian / Low Word First, High Byte First)
	ByteOrderBADC = "BADC" // Byte Swap (Mid-Big Endian / High Word First, Low Byte First)
	ByteOrderDCBA = "DCBA" // Little Endian (Full Reverse / Low Word First, Low Byte First)

	// 64-bit presets
	ByteOrderABCDEFGH = "ABCDEFGH" // Big Endian (Standard Modbus 64-bit)
	ByteOrderGHEFCDAB = "GHEFCDAB" // Word Swap (16-bit Word Reverse)
	ByteOrderBADCFEHG = "BADCFEHG" // Byte Swap within Words
	ByteOrderHGFEDCBA = "HGFEDCBA" // Little Endian (Full Reverse)
	ByteOrderEFGHABCD = "EFGHABCD" // Double Word Swap (32-bit DWord Swap)

	// Custom prefix marker
	CustomByteOrderPrefix = "CUSTOM:"
)

// DataTypeByteSize returns the required byte count for decoding the given data type
func DataTypeByteSize(dt model.ParameterDataType) int {
	switch dt {
	case model.DataTypeInt16, model.DataTypeUInt16:
		return 2
	case model.DataTypeFloat32, model.DataTypeInt32, model.DataTypeUInt32:
		return 4
	case model.DataTypeFloat64:
		return 8
	case model.DataTypeBoolean:
		return 2 // for register word; coil is handled separately as 1 bit/byte
	case model.DataTypeString:
		return 0 // variable length
	default:
		return 4
	}
}

// DefaultByteOrderForDataType returns the standard Big-Endian preset for a data type
func DefaultByteOrderForDataType(dt model.ParameterDataType) string {
	switch dt {
	case model.DataTypeInt16, model.DataTypeUInt16:
		return ByteOrderAB
	case model.DataTypeFloat64:
		return ByteOrderABCDEFGH
	case model.DataTypeBoolean, model.DataTypeString:
		return ByteOrderAB
	default:
		return ByteOrderABCD
	}
}

// ParsedByteOrder holds validated permutation details
type ParsedByteOrder struct {
	RawOrder     string // original input string
	Canonical    string // normalized permutation string, e.g. "CDAB" or "EFGHABCD"
	Permutation  []int  // slice of source byte indices for each output position
	IsCustom     bool   // true if explicitly configured as custom order
	ByteSize     int    // required byte size
}

// ParseAndValidateByteOrder parses and strictly validates a byte order string against the target data type
func ParseAndValidateByteOrder(order string, dataType model.ParameterDataType) (*ParsedByteOrder, error) {
	byteSize := DataTypeByteSize(dataType)
	if byteSize == 0 {
		// String or variable length: pass-through or optional word swap
		return &ParsedByteOrder{
			RawOrder:    order,
			Canonical:   order,
			Permutation: nil,
			IsCustom:    false,
			ByteSize:    0,
		}, nil
	}

	trimmed := strings.TrimSpace(order)
	isCustom := false
	if strings.HasPrefix(strings.ToUpper(trimmed), CustomByteOrderPrefix) {
		isCustom = true
		trimmed = strings.TrimSpace(trimmed[len(CustomByteOrderPrefix):])
	}

	// Clean order: strip spaces, commas, hyphens, underscores
	clean := strings.ToUpper(trimmed)
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, ",", "")
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, "_", "")

	// 1. Handle empty: default to standard Big-Endian for the data type
	if clean == "" {
		def := DefaultByteOrderForDataType(dataType)
		p, err := buildPermutation(def, byteSize)
		if err != nil {
			return nil, err
		}
		return &ParsedByteOrder{
			RawOrder:    order,
			Canonical:   def,
			Permutation: p,
			IsCustom:    false,
			ByteSize:    byteSize,
		}, nil
	}

	// 2. Handle standard legacy aliases
	switch clean {
	case "BIGENDIAN", "BIG_ENDIAN", "BE":
		clean = DefaultByteOrderForDataType(dataType)
	case "WORDSWAP", "WORD_SWAP", "WS":
		if byteSize == 8 {
			clean = ByteOrderGHEFCDAB
		} else {
			clean = ByteOrderCDAB
		}
	case "BYTESWAP", "BYTE_SWAP", "BS":
		if byteSize == 2 {
			clean = ByteOrderBA
		} else if byteSize == 8 {
			clean = ByteOrderBADCFEHG
		} else {
			clean = ByteOrderBADC
		}
	case "LITTLEENDIAN", "LITTLE_ENDIAN", "LE":
		if byteSize == 2 {
			clean = ByteOrderBA
		} else if byteSize == 8 {
			clean = ByteOrderHGFEDCBA
		} else {
			clean = ByteOrderDCBA
		}
	}

	// 3. Backward Compatibility Heuristic for legacy parameters:
	// If a 16-bit parameter has a 4-letter order (e.g. ABCD from previous system default):
	if byteSize == 2 && len(clean) == 4 {
		switch clean {
		case ByteOrderABCD, ByteOrderCDAB:
			clean = ByteOrderAB
		case ByteOrderBADC, ByteOrderDCBA:
			clean = ByteOrderBA
		}
	}
	// If an 8-byte (64-bit) parameter has a 4-letter order from previous system default:
	if byteSize == 8 && len(clean) == 4 {
		switch clean {
		case ByteOrderABCD:
			clean = ByteOrderABCDEFGH
		case ByteOrderCDAB:
			clean = ByteOrderGHEFCDAB
		case ByteOrderBADC:
			clean = ByteOrderBADCFEHG
		case ByteOrderDCBA:
			clean = ByteOrderHGFEDCBA
		}
	}

	// 4. Strict Validation of permutation:
	if len(clean) != byteSize {
		return nil, fmt.Errorf("invalid byte order '%s': requires exactly %d byte positions for %s, got %d ('%s')",
			order, byteSize, dataType, len(clean), clean)
	}

	maxChar := byte('A' + byteSize - 1)
	validChars := make([]string, byteSize)
	for i := 0; i < byteSize; i++ {
		validChars[i] = string(rune('A' + i))
	}
	validRangeStr := strings.Join(validChars, ", ")

	seen := make([]bool, byteSize)
	permutation := make([]int, byteSize)

	for i := 0; i < byteSize; i++ {
		ch := clean[i]
		if ch < 'A' || ch > maxChar {
			return nil, fmt.Errorf("invalid byte position '%c' in '%s': valid positions for %s are [%s]",
				ch, order, dataType, validRangeStr)
		}
		idx := int(ch - 'A')
		if seen[idx] {
			return nil, fmt.Errorf("invalid byte order '%s': duplicate byte position '%c' detected",
				order, ch)
		}
		seen[idx] = true
		permutation[i] = idx
	}

	for i := 0; i < byteSize; i++ {
		if !seen[i] {
			return nil, fmt.Errorf("invalid byte order '%s': missing byte position '%c' for %s",
				order, rune('A'+i), dataType)
		}
	}

	return &ParsedByteOrder{
		RawOrder:    order,
		Canonical:   clean,
		Permutation: permutation,
		IsCustom:    isCustom,
		ByteSize:    byteSize,
	}, nil
}

func buildPermutation(order string, byteSize int) ([]int, error) {
	p := make([]int, byteSize)
	for i := 0; i < byteSize; i++ {
		p[i] = int(order[i] - 'A')
	}
	return p, nil
}

// ValidateByteOrder verifies whether a byte order string is syntactically and semantically valid for a data type
func ValidateByteOrder(order string, dataType model.ParameterDataType) error {
	_, err := ParseAndValidateByteOrder(order, dataType)
	return err
}

// PermuteBytes rearranges raw input bytes according to the given permutation indices
func PermuteBytes(raw []byte, permutation []int) ([]byte, error) {
	n := len(permutation)
	if len(raw) < n {
		return nil, fmt.Errorf("insufficient raw bytes for permutation: need %d, got %d", n, len(raw))
	}

	out := make([]byte, n)
	for outIdx, srcIdx := range permutation {
		if srcIdx < 0 || srcIdx >= len(raw) {
			return nil, fmt.Errorf("source byte index %d out of bounds (len=%d)", srcIdx, len(raw))
		}
		out[outIdx] = raw[srcIdx]
	}
	return out, nil
}
