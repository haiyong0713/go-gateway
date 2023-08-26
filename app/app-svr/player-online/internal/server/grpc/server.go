package grpc

import (
	"context"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/library/net/rpc/warden"

	v1 "go-gateway/app/app-svr/player-online/api"
	"go-gateway/app/app-svr/player-online/internal/conf"
	"go-gateway/app/app-svr/player-online/internal/service/online"
)

type Server struct {
	onlineService *online.Service
	c             *conf.Config
}

func New(cfg *warden.ServerConfig, svr *online.Service, c *conf.Config) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(cfg)
	v1.RegisterPlayerOnlineServer(wsvr.Server(), &Server{onlineService: svr, c: c})
	// 用户鉴权
	auther := mauth.New(nil)
	wsvr.Add("/bilibili.app.playeronline.v1.PlayerOnline/PlayerOnline", auther.UnaryServerInterceptor(true))
	wsvr.Add("/bilibili.app.playeronline.v1.PlayerOnline/PremiereInfo", auther.UnaryServerInterceptor(true))
	wsvr.Add("/bilibili.app.playeronline.v1.PlayerOnline/ReportWatch", auther.UnaryServerInterceptor(true))
	wsvr, err = wsvr.Start()
	return
}

func (s *Server) PlayerOnline(c context.Context, req *v1.PlayerOnlineReq) (*v1.PlayerOnlineReply, error) {
	return s.onlineService.PlayerOnlineGRPC(c, req)
}

func (s *Server) PremiereInfo(c context.Context, req *v1.PremiereInfoReq) (*v1.PremiereInfoReply, error) {
	return s.onlineService.PremiereInfoGRPC(c, req)
}

func (s *Server) ReportWatch(c context.Context, req *v1.ReportWatchReq) (*v1.NoReply, error) {
	return s.onlineService.ReportWatchGRPC(c, req)
}
