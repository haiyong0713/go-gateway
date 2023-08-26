package tool

import (
	"math"
)

//Int64Append: 数字拼接,
//BenchmarkInt64Append-8   	 3404996	       322 ns/op
/*
	assert.Equal(t, int64(123456), Int64Append(123, 456))
	assert.Equal(t, int64(-123456), Int64Append(-123, 456))
	assert.Equal(t, int64(-123456), Int64Append(-123, -456))
*/
func Int64Append(a, b int64) (result int64) {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	var negative bool
	if a < 0 {
		negative = true
		a = -a
	}
	if b < 0 {
		negative = true
		b = -b
	}
	result = a
	times := int(math.Log10(float64(b)))
	for b != 0 {
		if result >= math.MaxInt64/10 {
			return -1
		}
		result = result * 10
		h := b / int64(math.Pow10(times))
		result = result + h
		b = b - (h * int64(math.Pow10(times)))
		times--
	}
	if negative {
		result = -result
	}
	return
}
