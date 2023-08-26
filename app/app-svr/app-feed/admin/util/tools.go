package util

import (
	"errors"
	"fmt"
	"sort"
)

func StringArrayDedup(in []string) (out []string) {
	tmp := map[string]bool{}
	for _, s := range in {
		if _, ok := tmp[s]; !ok {
			tmp[s] = true
			out = append(out, s)
		}
	}
	return
}

func Int64ArrayDedup(in []int64) (out []int64) {
	tmp := map[int64]bool{}
	for _, s := range in {

		if _, ok := tmp[s]; !ok {
			tmp[s] = true
			out = append(out, s)
		}
	}
	return
}

// Int2AlphaString converts int slice to num-alphabetic string.
// Rule:
//
//	number in [1, 9]:	return itself
//	number in [10, 35]:	convert int to alphabetic char by int2Char
//	number > 35:		return error
func Int2AlphaString(in []int) (out string, err error) {
	if len(in) == 0 {
		return "_", nil
	}
	sort.Ints(in)

	for _, i := range in {
		//nolint:gomnd
		if i < 0 {
			err = errors.New("Int2AlphaString: number must be positive")
			break
		} else if i < 10 {
			out = fmt.Sprintf("%s%d", out, i)
		} else if i < 36 {
			out = fmt.Sprintf("%s%s", out, int2Char(i-9))
		} else {
			err = errors.New("Int2AlphaString: number must be in range [1,35]")
			break
		}
	}
	return
}

func int2Char(i int) string {
	const dict = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return dict[i-1 : i]
}

func PaginateSlice(pn int, ps int, sliceLen int) (start int, end int) {
	if pn <= 0 || ps <= 0 {
		return
	}

	start = (pn - 1) * ps
	if start > sliceLen {
		start = sliceLen
	}

	end = start + ps
	if end > sliceLen {
		end = sliceLen
	}
	return
}

// 判断元素是否在数组中
func IsIntInArray(items []int, item int) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

// IsStringInArray 判断string元素是否在数组中
func IsStringInArray(items []string, item string) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}
