package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/model/entrance"
)

func entranceInfoc(ctx *bm.Context) {
	req := &entrance.BusinessInfocReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if midInter, ok := ctx.Get("mid"); ok {
		req.Mid = midInter.(int64)
	}
	ctx.JSON(nil, entranceSvc.BusinessInfoc(ctx, req))
}
