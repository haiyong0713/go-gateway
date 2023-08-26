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

type grpcServer struct{}

func (grpcServer) listDynPath(ctx *bm.Context) {
	req := &pb.ListDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	reply, err := rawSvc.GRPC.ListDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply.DynPaths, nil)
}

func parseAnnotation(dst *pb.SetDynPathReq, req *http.Request) error {
	annotation := req.Form.Get("annotation")
	if annotation == "" {
		return nil
	}
	dst.Annotation = make(map[string]string)
	if err := json.Unmarshal([]byte(annotation), &dst.Annotation); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (grpcServer) addDynPath(ctx *bm.Context) {
	req := &pb.SetDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseClientInfo(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	if err := parseAnnotation(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.AddDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (grpcServer) deleteDynPath(ctx *bm.Context) {
	req := &pb.DeleteDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.DeleteDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (grpcServer) updateDynPath(ctx *bm.Context) {
	req := &pb.SetDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseClientInfo(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	if err := parseAnnotation(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.UpdateDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (grpcServer) enableDynPath(ctx *bm.Context) {
	req := &pb.EnableDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.EnableDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (grpcServer) disableDynPath(ctx *bm.Context) {
	req := &pb.EnableDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.GRPC.DisableDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
