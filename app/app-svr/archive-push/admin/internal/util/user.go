package util

import (
	"context"
	"google.golang.org/grpc/metadata"
	"strconv"

	bm "go-common/library/net/http/blademaster"
)

func UserInfo(ctx context.Context) (username string, uid int64) {
	if bmContext, ok := ctx.(*bm.Context); ok {
		if nameInter, ok := bmContext.Get("username"); ok {
			username = nameInter.(string)
		}
		if uidInter, ok := bmContext.Get("uid"); ok {
			uid = uidInter.(int64)
		}
		if username == "" {
			cookie, _ := bmContext.Request.Cookie("username")
			if cookie == nil || cookie.Value == "" {
				return
			}
			username = cookie.Value
		}
		if uid == 0 {
			cookie, _ := bmContext.Request.Cookie("uid")
			if cookie == nil || cookie.Value == "" {
				return
			}
			uidInt, _ := strconv.Atoi(cookie.Value)
			uid = int64(uidInt)
		}
		return
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vusername := md.Get("username"); len(vusername) > 0 {
			username = vusername[0]
		}
		if vuid := md.Get("uid"); len(vuid) > 0 {
			uid, _ = strconv.ParseInt(vuid[0], 10, 64)
		}
	}
	return
}
