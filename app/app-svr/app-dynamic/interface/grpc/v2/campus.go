package v2

import (
	"context"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	mrestrict "go-common/component/restriction/middleware/grpc"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	apiV2 "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/http"
	dynsvrV2 "go-gateway/app/app-svr/app-dynamic/interface/service/dynamicV2"
	"go-gateway/app/app-svr/archive/middleware"
	middlewarev1 "go-gateway/app/app-svr/archive/middleware/v1"
)

type CampusServer struct {
	dynSvr *dynsvrV2.Service
	mauth  *mauth.Auth
	config *conf.Config
}

func InitCampusSvr(wd *warden.Server, auth *mauth.Auth, http *http.Server) (*CampusServer, error) {
	s := &CampusServer{
		dynSvr: http.DynamicSvcV2,
		mauth:  auth,
		config: http.Config,
	}
	apiV2.RegisterCampusServer(wd.Server(), s)

	wd.Add("/bilibili.app.dynamic.v2.Campus/WaterFlowRcmd", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), mrestrict.UnaryServerInterceptor())

	return s, nil
}

func (s *CampusServer) buildPlayerArgs(ctx context.Context, playArg *middlewarev1.PlayerArgs) context.Context {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取网络信息
	net, _ := network.FromContext(ctx)

	return middleware.NewContext(ctx, middleware.MossBatchPlayArgs(playArg, dev, net, au.Mid))
}
