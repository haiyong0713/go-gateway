package grpc

import (
	"context"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/space"
	"go-gateway/app/app-svr/app-interface/interface-legacy/http"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	spacesvc "go-gateway/app/app-svr/app-interface/interface-legacy/service/space"
	arcmid "go-gateway/app/app-svr/archive/middleware"
)

type SpaceServer struct {
	spaceSvc *spacesvc.Service
}

// nolint:unparam
func newSpace(ws *warden.Server, svr *http.Server) error {
	s := &SpaceServer{
		spaceSvc: svr.SpaceSvr,
	}
	api.RegisterSpaceServer(ws.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	ws.Add("/bilibili.app.interface.v1.Space/SearchTab", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	ws.Add("/bilibili.app.interface.v1.Space/SearchArchive", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	ws.Add("/bilibili.app.interface.v1.Space/SearchDynamic", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	return nil
}

func (s *SpaceServer) SearchTab(ctx context.Context, req *api.SearchTabReq) (*api.SearchTabReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	dev, _ := device.FromContext(ctx)
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	isIpad := model.IsIPad(plat)
	return s.spaceSvc.SearchTab(ctx, au.Mid, req, isIpad)
}

func (s *SpaceServer) SearchArchive(ctx context.Context, req *api.SearchArchiveReq) (*api.SearchArchiveReply, error) {
	au, _ := auth.FromContext(ctx)
	dev, _ := device.FromContext(ctx)
	network, _ := network.FromContext(ctx)
	batchArg := arcmid.MossBatchPlayArgs(req.PlayerArgs, dev, network, au.Mid)
	ctx = arcmid.NewContext(ctx, batchArg)
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	isIpad := model.IsIPad(plat)
	return s.spaceSvc.SearchArchive(ctx, req, isIpad)
}

func (s *SpaceServer) SearchDynamic(ctx context.Context, req *api.SearchDynamicReq) (*api.SearchDynamicReply, error) {
	au, _ := auth.FromContext(ctx)
	dev, _ := device.FromContext(ctx)
	net, _ := network.FromContext(ctx)
	ip := metadata.String(ctx, metadata.RemoteIP)
	return s.spaceSvc.SearchDynamic(ctx, au.Mid, req, dev, ip, net)
}
