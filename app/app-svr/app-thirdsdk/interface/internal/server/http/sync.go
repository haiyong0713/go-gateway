package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-thirdsdk/interface/internal/model"
)

func userBindSync(ctx *bm.Context) {
	param := &model.UserBindParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	ctx.JSON(nil, svc.UserBindSync(ctx, param, ip))
}

func arcStatusSync(ctx *bm.Context) {
	param := &model.ArcStatusParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	ctx.JSON(nil, svc.ArcStatusSync(ctx, param, ip))
}
