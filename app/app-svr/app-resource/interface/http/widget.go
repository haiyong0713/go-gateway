package http

import (
	bm "go-common/library/net/http/blademaster"
	model "go-gateway/app/app-svr/app-resource/interface/model/widget"
)

func widgets(ctx *bm.Context) {
	req := &model.WidgetsMetaReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(widgetSvc.WidgetMeta(ctx, req))
}

func widgetAndroid(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid, _ = midInter.(int64)
	}
	ctx.JSON(widgetSvc.WidgetAndroid(ctx, mid))
}
