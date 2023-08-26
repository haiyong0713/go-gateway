package http

import (
	bm "go-common/library/net/http/blademaster"
	appStoremdl "go-gateway/app/web-svr/activity/interface/model/appstore"
	"go-gateway/app/web-svr/activity/interface/service"
)

func appStoreState(ctx *bm.Context) {
	ua := ctx.Request.Header.Get("User-Agent")
	arg := new(appStoremdl.AppStoreStateArg)
	if err := ctx.Bind(arg); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	arg.MID = midStr.(int64)
	ctx.JSON(service.AppstoreSvc.AppStoreState(ctx, arg, ua))
}

func appStoreReceive(ctx *bm.Context) {
	ua := ctx.Request.UserAgent()
	path := ctx.Request.URL.Path
	referer := ctx.Request.Referer()
	arg := new(appStoremdl.AppStoreReceiveArg)
	if err := ctx.Bind(arg); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	arg.MID = midStr.(int64)
	ctx.JSON(nil, service.AppstoreSvc.APPStoreReceive(ctx, arg, ua, path, referer))
}
