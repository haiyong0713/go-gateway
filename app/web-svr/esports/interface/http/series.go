package http

import (
	bm "go-common/library/net/http/blademaster"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
)

func getSeriesPointMatchInfo(ctx *bm.Context) {
	v := &v1.GetSeriesPointMatchInfoReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(eSvc.HttpSeriesPointMatchInfo(ctx, v))
}

func getSeriesKnockoutMatchInfo(ctx *bm.Context) {
	var mid int64
	if d, ok := ctx.Get("mid"); ok {
		mid = d.(int64)
	}
	v := &v1.GetSeriesKnockoutMatchInfoReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(eSvc.GetSeriesKnockoutMatchInfoHttp(ctx, mid, v))
}
