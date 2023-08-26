package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"

	"go-gateway/app/app-svr/feature/service/api"
	"go-gateway/app/app-svr/feature/service/conf"
	"go-gateway/app/app-svr/feature/service/service"
)

type server struct {
	srv *service.Service
	c   *conf.Config
}

// New grpc server
func New(cfg *warden.ServerConfig, srv *service.Service, c *conf.Config) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(cfg)
	api.RegisterFeatureServer(wsvr.Server(), &server{srv: srv, c: c})
	wsvr, err = wsvr.Start()
	return
}

func (s *server) BuildLimit(c context.Context, req *api.BuildLimitReq) (res *api.BuildLimitReply, err error) {
	res, err = s.srv.BuildLimit(c, req)
	return
}

func (s *server) FeatureDegrades(c context.Context, req *api.FeatureDegradesReq) (res *api.FeatureDegradesReply, err error) {
	return s.srv.FeatureDegrades(c, req)
}

func (s *server) ChannelFeature(ctx context.Context, req *api.ChannelFeatureReq) (*api.ChannelFeatureReply, error) {
	return s.srv.ChannelFeature(ctx, req)
}

func (s *server) FeatureTVSwitch(c context.Context, req *api.FeatureTVSwitchReq) (*api.FeatureTVSwitchReply, error) {
	return s.srv.FeatureTVSwitch(c, req)
}

func (s *server) BusinessConfig(c context.Context, req *api.BusinessConfigReq) (*api.BusinessConfigReply, error) {
	return s.srv.BusinessConfig(c, req)
}

func (s *server) ABTest(c context.Context, req *api.ABTestReq) (res *api.ABTestReply, err error) {
	return s.srv.ABTest(c, req)
}
