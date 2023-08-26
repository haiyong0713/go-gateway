package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-channel/interface/model/pedia"
)

func baikeNav(ctx *bm.Context) {
	var params = &pedia.NavReq{}
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(channelSvcV2.BaikeNav(ctx, params))
}

func baikeFeed(ctx *bm.Context) {
	var params = &pedia.FeedReq{}
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(channelSvcV2.BaikeFeed(ctx, params))
}
