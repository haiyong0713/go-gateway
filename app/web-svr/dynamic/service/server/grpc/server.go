package grpc

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	api "go-gateway/app/web-svr/dynamic/service/api/v1"
	"go-gateway/app/web-svr/dynamic/service/service"
)

type server struct {
	srv *service.Service
}

// New Coin warden rpc server .
func New(c *warden.ServerConfig, svr *service.Service) (ws *warden.Server) {
	var (
		err error
	)
	ws = warden.NewServer(c)
	api.RegisterDynamicServer(ws.Server(), &server{srv: svr})
	if ws, err = ws.Start(); err != nil {
		panic(err)
	}
	return ws
}

// RegionTotal dynamic number (include live) .
func (s *server) RegionTotal(ctx context.Context, req *api.NoArgRequest) (res *api.RegionTotalReply, err error) {
	res = new(api.RegionTotalReply)
	tmp := s.srv.GoRegionTotal(ctx)
	res.Res = make(map[string]int64, len(tmp))
	for k, v := range tmp {
		res.Res[k] = int64(v)
	}
	return
}

// RegionArcs3 get region dynamic .
func (s *server) RegionArcs3(ctx context.Context, req *api.RegionArcs3Req) (*api.RegionArcs3Reply, error) {
	arcs, count, err := s.srv.DynamicRegion(ctx, req.Business, req.Rid, req.Pn, req.Ps, req.IsFilter)
	if err != nil {
		return nil, err
	}
	return &api.RegionArcs3Reply{
		Arcs:  arcs,
		Count: count,
	}, nil
}

// RegionTagArcs3 get region hot tag dynamic .
func (s *server) RegionTagArcs3(ctx context.Context, req *api.RegionTagArcs3Req) (res *api.RegionTagArcs3Reply, err error) {
	var count int
	res = new(api.RegionTagArcs3Reply)
	if res.Arcs, count, err = s.srv.GoRegionTagArcs3(ctx, int32(req.Rid), req.TagId, int(req.Pn), int(req.Ps)); err != nil {
		log.Error("[RegionTagArcs3] grpc RegionTagArcs3() rid(%d) tagid(%d) error(%v)", req.Rid, req.TagId, err)
		return
	}
	res.Count = int64(count)
	return
}

// RegAllArcs get region archive .
func (s *server) RegAllArcs(ctx context.Context, req *api.RegAllReq) (res *api.RegAllReply, err error) {
	res = new(api.RegAllReply)
	if req.Type == 0 {
		if res.Archives, res.Count, err = s.srv.RegAllArc(ctx, req.Rid, int(req.Ps), int(req.Pn)); err != nil {
			log.Error("[RegAllArc] grpc RegAllArcs() error(%v)", err)
		}
	} else {
		if res.Archives, res.Count, err = s.srv.RegOriginArc(ctx, req.Rid, int(req.Ps), int(req.Pn)); err != nil {
			log.Error("[RegOriginArc] grpc RegAllArcs() error(%v)", err)
		}
	}
	return
}

// RegCount get second region one day archive count .
func (s *server) RegCount(ctx context.Context, req *api.RegCountReq) (res *api.RegCountReply, err error) {
	res = new(api.RegCountReply)
	if len(req.Rid) == 0 {
		return
	}
	if res.RegCountMap, err = s.srv.RegionCnt(ctx, req.Rid); err != nil {
		log.Error("[RegCount] grpc RegionCnt() rids(%v) error(%v)", req.Rid, err)
	}
	return
}

// RecentThrdRegArc 获取分区最近三天稿件
func (s *server) RecentThrdRegArc(ctx context.Context, req *api.RecentThrdRegArcReq) (res *api.RecentThrdRegArcReply, err error) {
	res = new(api.RecentThrdRegArcReply)
	if res.Archives, err = s.srv.RecentThrdRegArc(ctx, req.Rid, int(req.Pn), int(req.Ps)); err != nil {
		log.Error("[RecentThrdRegArc] grpc RecentThrdRegArc() rid(%d) pn(%d) ps(%d) error(%v)", req.Rid, req.Pn, req.Ps, err)
	}
	return
}

// RecentWeeklyArc 获取最近七天稿件
func (s *server) RecentWeeklyArc(ctx context.Context, req *api.RecentWeeklyArcReq) (res *api.RecentWeeklyArcReply, err error) {
	res = new(api.RecentWeeklyArcReply)
	if res.Archives, res.Count, err = s.srv.RecentWeeklyArc(ctx, int(req.Pn), int(req.Ps)); err != nil {
		log.Error("[RecentWeeklyArc] grpc RecentWeeklyArc() pn(%d) ps(%d) error(%v)", req.Pn, req.Ps, err)
	}
	return
}
