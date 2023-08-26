package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"
)

func appPromptAPI(ctx *bm.Context) {
	req := &api.AppPromptAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	req.Cookie = ctx.Request.Header.Get("Cookie")
	reply, err := rawSvc.Common.AppPromptAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply.Nodes, nil)
}

func configPromptAPI(ctx *bm.Context) {
	req := &api.ConfigPromptAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Cookie = ctx.Request.Header.Get("Cookie")
	res, err := rawSvc.Common.ConfigPromptAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func appPathPromptAPI(ctx *bm.Context) {
	req := &api.AppPathPromptAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	res, err := rawSvc.Common.AppPathPromptAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func zonePromptAPI(ctx *bm.Context) {
	req := &api.ZonePromptAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(rawSvc.Common.ZonePromptAPI(ctx, req))
}

func grpcAppMethodPromptAPI(ctx *bm.Context) {
	req := &api.AppPathPromptAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(rawSvc.Common.GRPCAppMethodPromptAPI(ctx, req))
}

func grpcAppPackagePromptAPI(ctx *bm.Context) {
	req := &api.GRPCAppPackagePromptAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	svrMd, err := rawSvc.Common.GRPCAppPackagePromptAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, err)
	}
	ret := map[string][]string{}
	for pkg, svrApi := range svrMd.Package {
		ret[pkg] = append(ret[pkg], svrApi.Services...)
	}
	ctx.JSON(ret, nil)
}
