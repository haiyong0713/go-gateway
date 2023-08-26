package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// nolint:gomnd
func (s *Service) CaptchaKey(_ context.Context, key string, needJsRes bool) (string, string) {
	resKey := func() string {
		if len(key) == 32 {
			return key
		}
		chars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		var res string
		rand.Seed(time.Now().UnixNano())
		charLen := len(chars)
		for i := 0; i < charLen; i++ {
			randPos := rand.Intn(charLen)
			var buffer bytes.Buffer
			buffer.WriteString(res)
			buffer.WriteString(chars[randPos : randPos+1])
			res = buffer.String()
		}
		nowHour := time.Now().Unix() / 60 * 60
		mh := md5.Sum([]byte(fmt.Sprintf("%s%dbilibili", res, nowHour)))
		return hex.EncodeToString(mh[:])
	}()
	if needJsRes {
		return resKey, fmt.Sprintf(`window.captcha_key = "%s";`, resKey)
	}
	return resKey, fmt.Sprintf(`"%s"`, resKey)
}

func (s *Service) ServerDate(_ context.Context, needTsRes bool) string {
	now := time.Now()
	if needTsRes {
		_, offset := now.Zone()
		return strconv.FormatInt(now.Unix()-int64(offset), 10)
	}
	nowUTC := now.UTC()
	return fmt.Sprintf("window.serverdate = Date.UTC(%d, %d, %d, %d, %d, %d);", nowUTC.Year(), nowUTC.Month()-1, nowUTC.Day(), nowUTC.Hour(), nowUTC.Minute(), nowUTC.Second())
}
