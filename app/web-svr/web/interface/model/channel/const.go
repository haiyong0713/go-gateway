package channel

import (
	"strconv"
	"strings"
)

const (
	CardTypeVideo  = 0
	CardTypeCustom = 256
	CardTypeRank   = 257
	CTypeNew       = 2
	VideoChannel   = 3
	CategoryHot    = 100

	_ogvIconURL = "https://i0.hdslb.com/bfs/tag/dd3b2d6991e3dd8640a5412282f1e08800d1e0b6.png"
	RankURL     = "https://www.bilibili.com/h5/channel/rank?id=%v&theme=%v&navhide=1"
)

// StatString Stat to string
// nolint:gomnd
func StatString(number int32, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(int64(number), 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

// StatString Stat to string
// nolint:gomnd
func Stat64String(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}
