package function

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

func Now() int64 {
	return time.Now().Unix()
}

func InInt64Slice(find int64, set []int64) bool {
	for _, v := range set {
		if find == v {
			return true
		}
	}
	return false
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
