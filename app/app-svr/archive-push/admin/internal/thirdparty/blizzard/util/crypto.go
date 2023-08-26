package util

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
)

func Sign(req model.VodAddReq, key string) string {
	mac := hmac.New(md5.New, []byte(key))
	toSignStr := fmt.Sprintf("%s%s%s%d%d%s%d%s%s%d", req.BVID, req.Category, req.Description, req.Duration, req.Page, req.Stage, req.Status, req.Thumbnail, req.Title, req.Timestamp)
	mac.Write([]byte(toSignStr))
	ms := mac.Sum(nil)
	return hex.EncodeToString(ms[:])
}
