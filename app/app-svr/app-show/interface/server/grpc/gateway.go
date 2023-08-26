package grpc

import (
	"context"

	"go-gateway/app/app-svr/app-show/interface/http"
	"go-gateway/app/app-svr/app-show/interface/service/act"

	api "git.bilibili.co/bapis/bapis-go/bilibili/app/show/gateway/v1"
	mauth "go-common/component/auth/middleware/grpc"
	"go-common/library/net/rpc/warden"
)

type GatewayServer struct {
	actSvc *act.Service
}

func gatewayGRPC(wsvr *warden.Server, svr *http.Server) {
	s := &GatewayServer{
		actSvc: svr.ActSvr,
	}
	api.RegisterAppShowServer(wsvr.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	wsvr.Add("/bilibili.app.show.v1.AppShow/GetActProgress", auther.UnaryServerInterceptor(true))
}

func (s *GatewayServer) GetActProgress(c context.Context, req *api.GetActProgressReq) (*api.GetActProgressReply, error) {
	return s.actSvc.GetActProgress(c, req)
}
