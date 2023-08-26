package http

import (
	"encoding/json"
	"net/http"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"

	"github.com/pkg/errors"
)

func listLimiter(ctx *bm.Context) {
	req := &pb.ListLimiterReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	reply, err := rawSvc.HTTP.ListLimiter(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply.List, nil)
}

func addLimiter(ctx *bm.Context) {
	req := &pb.AddLimiterReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.AddLimiter(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func updateLimiter(ctx *bm.Context) {
	req := &pb.SetLimiterReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.UpdateLimiter(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func deleteLimiter(ctx *bm.Context) {
	req := &pb.DeleteLimiterReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.DeleteLimiter(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func enableLimiter(ctx *bm.Context) {
	req := &pb.EnableLimiterReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.EnableLimiter(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func disableLimiter(ctx *bm.Context) {
	req := &pb.EnableLimiterReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.DisableLimiter(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func setupPlugin(ctx *bm.Context) {
	req := &pb.PluginReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parsePlugin(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	res, err := rawSvc.HTTP.SetupPlugin(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func parsePlugin(dst *pb.PluginReq, req *http.Request) error {
	plugin := req.Form.Get("plugin")
	if plugin == "" {
		return nil
	}
	dst.Plugin = &pb.Plugin{}
	if err := json.Unmarshal([]byte(plugin), dst.Plugin); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func pluginList(ctx *bm.Context) {
	req := &pb.PluginListReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	res, err := rawSvc.HTTP.PluginList(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
