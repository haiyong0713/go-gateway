package util

import (
	"strconv"
	"strings"
)

func Int64ArrayIn(list []int64, target int64) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func Int64ArrayJoin(list []int64, sep string) string {
	var strs []string
	for _, v := range list {
		strs = append(strs, strconv.FormatInt(v, 10))
	}
	return strings.Join(strs, ",")
}

func Int32ArrayIn(list []int32, target int32) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
