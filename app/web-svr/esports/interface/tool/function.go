package tool

import (
	"math"
)

func Unique(ids []int64) (outs []int64) {
	idMap := make(map[int64]int64, len(ids))
	for _, v := range ids {
		if _, ok := idMap[v]; ok {
			continue
		} else {
			idMap[v] = v
		}
		outs = append(outs, v)
	}
	return
}

func DecimalFloat(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}
