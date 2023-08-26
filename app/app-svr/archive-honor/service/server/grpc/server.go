package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"

	"go-gateway/app/app-svr/archive-honor/service/api"
	"go-gateway/app/app-svr/archive-honor/service/conf"
	"go-gateway/app/app-svr/archive-honor/service/service"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

type server struct {
	srv *service.Service
	c   *conf.Config
}

// New grpc server
func New(cfg *warden.ServerConfig, srv *service.Service, c *conf.Config) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(cfg)
	api.RegisterArchiveHonorServer(wsvr.Server(), &server{srv: srv, c: c})
	wsvr, err = wsvr.Start()
	return
}

func (s *server) Honor(c context.Context, req *api.HonorRequest) (resp *api.HonorReply, err error) {
	resp = new(api.HonorReply)
	ctx := s.srv.Feature.BuildLimitManual(c, &feature.BuildLimitManual{
		Build:   req.Build,
		MobiApp: req.MobiApp,
		Device:  req.Device,
	})
	resp.Honor, err = s.srv.Honor(ctx, req.Aid)
	return
}

func (s *server) Honors(c context.Context, req *api.HonorsRequest) (resp *api.HonorsReply, err error) {
	resp = new(api.HonorsReply)
	ctx := s.srv.Feature.BuildLimitManual(c, &feature.BuildLimitManual{
		Build:   req.Build,
		MobiApp: req.MobiApp,
		Device:  req.Device,
	})
	resp.Honors, err = s.srv.Honors(ctx, req.Aids)
	return
}
