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

func (httpServer) listDynPath(ctx *bm.Context) {
	req := &pb.ListDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	reply, err := rawSvc.HTTP.ListDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply.DynPaths, nil)
}

func parseClientInfo(dst *pb.SetDynPathReq, req *http.Request) error {
	clientInfo := req.Form.Get("client_info")
	if clientInfo == "" {
		return nil
	}
	dst.ClientInfo = &pb.ClientInfo{}
	if err := json.Unmarshal([]byte(clientInfo), dst.ClientInfo); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (httpServer) addDynPath(ctx *bm.Context) {
	req := &pb.SetDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseClientInfo(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.AddDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (httpServer) deleteDynPath(ctx *bm.Context) {
	req := &pb.DeleteDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.DeleteDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (httpServer) updateDynPath(ctx *bm.Context) {
	req := &pb.SetDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseClientInfo(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.UpdateDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (httpServer) enableDynPath(ctx *bm.Context) {
	req := &pb.EnableDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.EnableDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (httpServer) disableDynPath(ctx *bm.Context) {
	req := &pb.EnableDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.DisableDynPath(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
