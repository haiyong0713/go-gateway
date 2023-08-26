package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	brandMdl "go-gateway/app/web-svr/activity/interface/model/brand"
	"go-gateway/app/web-svr/activity/interface/service"
)

func coupon(ctx *bm.Context) {

	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ip := metadata.String(ctx, metadata.RemoteIP)
	ua := ctx.Request.UserAgent()
	path := ctx.Request.URL.Path
	referer := ctx.Request.Referer()
	device, err := ctx.Request.Cookie("buvid3")
	params := brandMdl.FrontEndParams{
		IP:      ip,
		Ua:      ua,
		API:     path,
		Referer: referer,
	}
	if err == nil {
		params.DeviceID = device.Value
	}
	ctx.JSON(service.BrandSvc.AddCoupon(ctx, midI, &params))
}
