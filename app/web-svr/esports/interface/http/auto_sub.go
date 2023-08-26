package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/esports/interface/model"
)

func autoSubscribeStatus(ctx *bm.Context) {
	var mid int64

	req := new(model.AutoSubRequest)
	if err := ctx.Bind(req); err != nil {
		return
	}

	if d, ok := ctx.Get("mid"); ok {
		mid = d.(int64)
	}

	d, err := eSvc.AutoSubscribeStatus(ctx, mid, req)
	ctx.JSON(d, err)
}

func autoSubscribe(ctx *bm.Context) {
	var mid int64

	req := new(model.AutoSubRequest)
	if err := ctx.Bind(req); err != nil {
		return
	}

	if d, ok := ctx.Get("mid"); ok {
		mid = d.(int64)
	}

	err := eSvc.AutoSubscribe(ctx, mid, req)
	ctx.JSON(nil, err)
}
