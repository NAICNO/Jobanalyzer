// Misc utilities for helping out with json encoding and decoding

package json

import (
	"math"
)

// Json can't represent NaN or Infinity (only one of the ways it's broken) so clamp a
// floating point value.
//
// Infinity  -> MaxFloat64
// -Infinity -> -MaxFloat64
// NaN       -> 0

func CleanFloat64(f float64) float64 {
	if math.IsInf(f, 1) {
		return math.MaxFloat64
	}
	if math.IsInf(f, -1) {
		return -math.MaxFloat64
	}
	if math.IsNaN(f) {
		return 0
	}
	return f
}
