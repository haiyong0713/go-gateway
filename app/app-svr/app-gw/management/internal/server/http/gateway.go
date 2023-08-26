package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"

	"github.com/pkg/errors"
)

func parseConfig(dst *pb.SetGatewayReq, req *http.Request) error {
	configs := req.Form.Get("configs")
	if configs == "" {
		return nil
	}
	dst.Configs = []*pb.ConfigMeta{}
	if err := json.Unmarshal([]byte(configs), &dst.Configs); err != nil {
		return errors.WithStack(err)
	}
	grpcConfigs := req.Form.Get("grpc_configs")
	if grpcConfigs == "" {
		return nil
	}
	dst.GrpcConfigs = []*pb.ConfigMeta{}
	if err := json.Unmarshal([]byte(grpcConfigs), &dst.GrpcConfigs); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func gateway(ctx *bm.Context) {
	req := &pb.AuthZReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Cookie = ctx.Request.Header.Get("Cookie")
	req.Username = username.(string)
	reply, err := rawSvc.Common.Gateway(ctx, req)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(reply.Gateways, nil)
}

func gatewayProfile(ctx *bm.Context) {
	req := &pb.GatewayProfileReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Cookie = ctx.Request.Header.Get("Cookie")
	req.Username = username.(string)
	req.Host = ctx.Request.Host
	reply, err := rawSvc.Common.GatewayProfile(ctx, req)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(reply, nil)
}

func addGateway(ctx *bm.Context) {
	req := &pb.SetGatewayReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseConfig(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Common.AddGateway(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func updateGateway(ctx *bm.Context) {
	req := &pb.SetGatewayReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseConfig(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Common.UpdateGateway(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func deleteGateway(ctx *bm.Context) {
	req := &pb.DeleteGatewayReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Common.DeleteGateway(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func enableAllGateway(ctx *bm.Context) {
	req := &pb.UpdateALLGatewayConfigReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Common.EnableALLGatewayConfig(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func disableAllGateway(ctx *bm.Context) {
	req := &pb.UpdateALLGatewayConfigReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Common.DisableALLGatewayConfig(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func gatewayProxy(ctx *bm.Context) {
	token := ctx.Params.ByName("token")
	strs := strings.Split(ctx.RoutePath, ":token")
	if len(strs) <= 1 {
		ctx.String(400, "Failed to split: %v, sep: %s", ctx.RoutePath, ":token")
		return
	}
	req := &pb.GatewayProxyReq{
		Token:  token,
		Suffix: path.Join("/", "_", strs[1]),
	}
	reply, err := rawSvc.Common.GatewayProxy(ctx, req)
	if err != nil {
		ctx.String(500, "%v", err)
		return
	}
	for k, v := range reply.Header {
		for _, val := range v.Values {
			ctx.Writer.Header().Add(k, val)
		}
	}
	ctx.Writer.WriteHeader(int(reply.StatusCode))
	if _, err := io.Copy(ctx.Writer, bytes.NewReader(reply.Page)); err != nil {
		ctx.String(500, "%v", errors.WithStack(err))
		return
	}
}

func initGatewayConfigs(ctx *bm.Context) {
	req := &pb.InitGatewayConfigsReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Cookie = ctx.Request.Header.Get("Cookie")
	res, err := rawSvc.Common.InitGatewayConfigs(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func enableAllGRPCGateway(ctx *bm.Context) {
	req := &pb.UpdateALLGatewayConfigReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Common.EnableAllGRPCGatewayConfig(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func disableAllGRPCGateway(ctx *bm.Context) {
	req := &pb.UpdateALLGatewayConfigReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Common.DisableAllGRPCGatewayConfig(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
