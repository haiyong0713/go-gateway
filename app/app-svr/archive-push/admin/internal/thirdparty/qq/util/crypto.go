package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func Sign(secret string, source string, ibiz string, timestamp int64) string {
	if secret == "" || source == "" || ibiz == "" {
		return ""
	}
	toSignStr := fmt.Sprintf("%s%s%s%d", secret, source, ibiz, timestamp)
	ms := md5.Sum([]byte(toSignStr))
	return hex.EncodeToString(ms[:])
}
