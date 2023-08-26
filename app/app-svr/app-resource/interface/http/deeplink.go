package http

import (
	bm "go-common/library/net/http/blademaster"
	model "go-gateway/app/app-svr/app-resource/interface/model/deeplink"
)

func deeplinkHW(ctx *bm.Context) {
	req := &model.HWDeeplinkReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(deeplinkSvc.DeepLinkHW(ctx, req))
	//nolint:gosimple
	return
}

func deeplinkButton(ctx *bm.Context) {
	req := &model.BackButtonReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(deeplinkSvc.BackButton(ctx, req))
	//nolint:gosimple
	return
}

func deeplinkAi(ctx *bm.Context) {
	req := &model.AiDeeplinkReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	header := ctx.Request.Header
	buvid := header.Get("Buvid")
	ctx.JSON(deeplinkSvc.DeepLinkAI(ctx, req, buvid))
}
