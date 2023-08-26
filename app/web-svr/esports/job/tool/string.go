package tool

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"
)

const (
	DelimiterOfComma = ","
)

func Int64JoinStr(elems []int64, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("%v", elems[0])
	}

	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(fmt.Sprintf("%v", elems[i]))
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(fmt.Sprintf("%v", elems[0]))
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(fmt.Sprintf("%v", s))
	}

	return b.String()
}

func Int64InSlice(a int64, list []int64) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StrToFloatNormal(s string) float64 {
	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Error("strToFloatNormal s(%s) error(%+v)", s, err)
		return 0
	}
	return floatValue
}

func StrToInt64Normal(s string) int64 {
	int64Value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Error("strToInt64Normal s(%s) error(%+v)", s, err)
		return 0
	}
	return int64Value
}
