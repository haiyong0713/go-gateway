package utils

import (
	"strconv"
	"strings"
)

func Sting2IntSlice(input string, splitter string) []int {
	strs := strings.Split(input, splitter)
	ary := make([]int, len(strs))
	for i := range ary {
		ary[i], _ = strconv.Atoi(strs[i])
	}
	return ary
}
