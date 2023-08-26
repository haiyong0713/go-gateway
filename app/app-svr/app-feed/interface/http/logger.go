package http

import (
	"fmt"
	"hash/crc32"
	"strconv"
)

//nolint:deadcode,unused
func aiAd(mid int64, buvid string) bool {
	var group int
	if mid > 0 {
		group = int(mid % 20)
	} else {
		group = int(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s_1CF61D5DE42C7852", buvid))) % 4)
	}
	if mid > 0 {
		if _, ok := cfg.Custom.AIAdMid[strconv.FormatInt(mid, 10)]; ok {
			return true
		}
		if _, ok := cfg.Custom.AIAdGroupMid[strconv.Itoa(group)]; ok {
			return true
		}
		return false
	}
	if _, ok := cfg.Custom.AIAdGroupBuvid[strconv.Itoa(group)]; ok {
		return true
	}
	return false
}
