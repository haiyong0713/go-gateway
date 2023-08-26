package grpc

import (
	"context"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/ugc-season/service/api"
	"go-gateway/app/app-svr/ugc-season/service/conf"
	"go-gateway/app/app-svr/ugc-season/service/service"
)

type server struct {
	srv *service.Service
	c   *conf.Config
}

// New grpc server
func New(cfg *warden.ServerConfig, srv *service.Service, c *conf.Config) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(cfg)
	api.RegisterUGCSeasonServer(wsvr.Server(), &server{srv: srv, c: c})
	wsvr, err = wsvr.Start()
	return
}

func (s *server) Season(c context.Context, req *api.SeasonRequest) (resp *api.SeasonReply, err error) {
	resp = new(api.SeasonReply)
	resp.Season, err = s.srv.Season(c, req.SeasonID)
	return
}

func (s *server) Seasons(c context.Context, req *api.SeasonsRequest) (resp *api.SeasonsReply, err error) {
	resp = new(api.SeasonsReply)
	resp.Seasons, err = s.srv.Seasons(c, req.SeasonIds)
	return
}

func (s *server) View(c context.Context, req *api.ViewRequest) (resp *api.ViewReply, err error) {
	resp = new(api.ViewReply)
	resp.View, err = s.srv.View(c, req.SeasonID)
	return
}

func (s *server) Views(c context.Context, req *api.ViewsRequest) (resp *api.ViewsReply, err error) {
	resp = new(api.ViewsReply)
	resp.Views, err = s.srv.Views(c, req.SeasonIds, req.EpSize)
	return
}

func (s *server) Stat(c context.Context, req *api.StatRequest) (resp *api.StatReply, err error) {
	resp = new(api.StatReply)
	resp.Stat, err = s.srv.Stat(c, req.SeasonID)
	return
}

func (s *server) Stats(c context.Context, req *api.StatsRequest) (resp *api.StatsReply, err error) {
	resp = new(api.StatsReply)
	resp.Stats = make(map[int64]*api.Stat)
	resp.Stats, err = s.srv.Stats(c, req.SeasonIDs)
	return
}

func (s *server) UpperList(c context.Context, req *api.UpperListRequest) (resp *api.UpperListReply, err error) {
	resp = new(api.UpperListReply)
	resp.Seasons, resp.TotalCount, resp.TotalPage, err = s.srv.UpperSeason(c, req)
	return
}

func (s *server) UpCache(c context.Context, req *api.UpCacheRequest) (resp *api.NoReply, err error) {
	resp = new(api.NoReply)
	err = s.srv.UpSeasonCache(c, req.SeasonID, req.Action)
	return
}
