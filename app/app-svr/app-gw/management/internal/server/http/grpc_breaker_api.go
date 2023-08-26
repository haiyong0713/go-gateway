package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"
)

func (grpcServer) listBreakerAPI(ctx *bm.Context) {
	req := &pb.ListBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	reply, err := rawSvc.GRPC.ListBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	out := make([]*model.BreakerAPI, 0, len(reply.BreakerApiList))
	for _, pbba := range reply.BreakerApiList {
		ba := &model.BreakerAPI{}
		ba.FromProto(pbba)
		out = append(out, ba)
	}
	ctx.JSON(out, nil)
}

func (grpcServer) addBreakerAPI(ctx *bm.Context) {
	req := &pb.SetBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseAction(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	if err := parseFlowCopy(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.SetBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (grpcServer) updateBreakerAPI(ctx *bm.Context) {
	req := &pb.SetBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseAction(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	if err := parseFlowCopy(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.UpdateBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (grpcServer) enableBreakerAPI(ctx *bm.Context) {
	req := &pb.EnableBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.EnableBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (grpcServer) disableBreakerAPI(ctx *bm.Context) {
	req := &pb.EnableBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.DisableBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (grpcServer) deleteBreakerAPI(ctx *bm.Context) {
	req := &pb.DeleteBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.DeleteBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
