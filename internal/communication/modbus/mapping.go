package modbus

import (
	"math"
)

// ApplyScaleAndOffset calculates the engineering value from the raw numeric register value
// Formula: engineering_value = (raw_value * scale) + offset
// If scale is 0, default multiplier 1.0 is applied.
// If precision >= 0, the result is rounded to that number of decimal places.
func ApplyScaleAndOffset(rawValue float64, scale, offset float64, precision int) float64 {
	factor := scale
	if factor == 0 {
		factor = 1.0
	}

	engVal := (rawValue * factor) + offset

	if precision >= 0 && precision <= 10 {
		pow := math.Pow10(precision)
		engVal = math.Round(engVal*pow) / pow
	}

	return engVal
}
