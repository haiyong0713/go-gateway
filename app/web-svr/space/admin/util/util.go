package util

import (
	"strconv"
	"strings"
)

func SplitInt(src string) (ret []int) {
	list := strings.Split(src, ",")
	ret = make([]int, 0, len(list))
	for _, s := range list {
		num, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		ret = append(ret, num)
	}
	return
}
