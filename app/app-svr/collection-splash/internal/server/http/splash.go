package http

import (
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/collection-splash/api"

	"github.com/golang/protobuf/ptypes/empty"
)

func addSplash(ctx *bm.Context) {
	req := &pb.AddSplashReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.AddSplash(ctx, req))
}

func updateSplash(ctx *bm.Context) {
	req := &pb.UpdateSplashReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.UpdateSplash(ctx, req))
}

func deleteSplash(ctx *bm.Context) {
	req := &pb.SplashReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.DeleteSplash(ctx, req))
}

func splash(ctx *bm.Context) {
	req := &pb.SplashReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.Splash(ctx, req))
}

func splashList(ctx *bm.Context) {
	ctx.JSON(svc.SplashList(ctx, &empty.Empty{}))
}
