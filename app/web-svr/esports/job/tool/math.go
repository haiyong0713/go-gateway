package tool

import "math"

// Calculate ln X math result
// e.g: ln 365 ~= 5.8998973535825
func LNX(input float64) float64 {
	return math.Log2(input) * math.Log2E
}
