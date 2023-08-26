package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"

	"go-gateway/app/app-svr/archive-extra/service/api"
	"go-gateway/app/app-svr/archive-extra/service/conf"
	"go-gateway/app/app-svr/archive-extra/service/service"
)

type server struct {
	srv *service.Service
	c   *conf.Config
}

func (s *server) AddArchiveExtraValue(ctx context.Context, req *api.AddArchiveExtraValueReq) (*api.NoReply, error) {
	err := s.srv.AddArchiveExtraValue(ctx, req.Aid, req.Key, req.Value)
	return &api.NoReply{}, err
}

func (s *server) BatchAddArchiveExtraValue(ctx context.Context, req *api.BatchAddArchiveExtraValueReq) (*api.NoReply, error) {
	err := s.srv.BatchAddArchiveExtraValue(ctx, req.Key, req.AidValues)
	return &api.NoReply{}, err
}

func (s *server) RemoveArchiveExtraValue(ctx context.Context, req *api.RemoveArchiveExtraValueReq) (*api.NoReply, error) {
	err := s.srv.RemoveArchiveExtraValue(ctx, req.Aid, req.Key)
	return &api.NoReply{}, err
}

func (s *server) BatchRemoveArchiveExtraValue(ctx context.Context, req *api.BatchRemoveArchiveExtraValueReq) (*api.NoReply, error) {
	err := s.srv.BatchRemoveArchiveExtraValue(ctx, req.Aids, req.Key)
	return &api.NoReply{}, err
}

func (s *server) GetArchiveExtraValue(ctx context.Context, req *api.GetArchiveExtraValueReq) (*api.ArchiveExtraValueReply, error) {
	resp := new(api.ArchiveExtraValueReply)
	extraInfo, err := s.srv.GetArchiveExtraValue(ctx, req.Aid)
	if err != nil {
		return nil, err
	}
	resp.ExtraInfo = extraInfo
	return resp, nil
}

func (s *server) BatchGetArchiveExtraValue(ctx context.Context, req *api.BatchGetArchiveExtraValueReq) (resp *api.MultiArchiveExtraValueReply, err error) {
	resp = new(api.MultiArchiveExtraValueReply)
	resp.ExtraInfos = make(map[int64]*api.ArchiveExtraValueReply)
	info, err := s.srv.BatchGetArchiveExtraValue(ctx, req.Aids)
	if err != nil {
		return
	}
	if len(info) == 0 {
		return
	}
	for aid, a := range info {
		resp.ExtraInfos[aid] = a
	}

	return resp, nil
}

func (s *server) GetArchiveExtraBasedOnKeys(ctx context.Context, req *api.GetArchiveExtraBasedOnKeysReq) (*api.ArchiveExtraValueReply, error) {
	resp := new(api.ArchiveExtraValueReply)
	extraInfo, err := s.srv.GetArchiveExtraBasedOnKeys(ctx, req.Aid, req.Keys)
	if err != nil {
		return nil, err
	}
	resp.ExtraInfo = extraInfo
	return resp, nil
}

// New grpc server
func New(cfg *warden.ServerConfig, srv *service.Service, c *conf.Config) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(cfg)
	api.RegisterArcExtraServer(wsvr.Server(), &server{srv: srv, c: c})
	wsvr, err = wsvr.Start()
	return
}
