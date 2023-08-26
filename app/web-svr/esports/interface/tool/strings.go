package tool

import (
	"strconv"
	"strings"
)

const (
	delimiterOfComma = ","
)

func SplitString2Int64(s, sep string) []int64 {
	list := make([]int64, 0)
	if d := strings.Split(s, sep); len(d) > 0 {
		for _, v := range d {
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				list = append(list, i)
			}
		}
	}

	return list
}

func Int64InSlice(a int64, list []int64) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
