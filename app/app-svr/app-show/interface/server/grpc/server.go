package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"

	pb "go-gateway/app/app-svr/app-show/interface/api"
	"go-gateway/app/app-svr/app-show/interface/http"
	actsvr "go-gateway/app/app-svr/app-show/interface/service/act"
	service "go-gateway/app/app-svr/app-show/interface/service/show"
)

// New History warden rpc server
func New(c *warden.ServerConfig, svr *http.Server) *warden.Server {
	ws := warden.NewServer(c)
	pb.RegisterAppShowServer(ws.Server(), &server{svr.ShowSvc, svr.ActSvr})
	// rank
	rankGRPC(ws, svr)
	PopularGRPC(ws, svr)
	gatewayGRPC(ws, svr)
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}

type server struct {
	svr    *service.Service
	actSvr *actsvr.Service
}

var _ pb.AppShowServer = &server{}

// RefreshSeriesList refreshes  the series list cache in redis
func (s *server) RefreshSeriesList(ctx context.Context, req *pb.RefreshSeriesListReq) (*pb.NoReply, error) {
	_, err := s.svr.BackToSrcSeries(ctx, req.Type)
	return &pb.NoReply{}, err
}

// RefreshSerie refreshes one serie with the given type and number
func (s *server) RefreshSerie(ctx context.Context, req *pb.RefreshSerieReq) (*pb.NoReply, error) {
	_, err := s.svr.BackToSrcSerie(ctx, req.Type, req.Number)
	return &pb.NoReply{}, err
}

// ActNativeTab --供动态测使用.
func (s *server) ActNativeTab(ctx context.Context, req *pb.ActNativeTabReq) (*pb.ActNativeTabReply, error) {
	rly, e := s.actSvr.ActNativeTab(ctx, req)
	if e != nil {
		return nil, e
	}
	if rly == nil {
		return &pb.ActNativeTabReply{}, nil
	}
	return rly, nil
}

// ActShare --供分享组件使用.
func (s *server) ActShare(ctx context.Context, req *pb.ActShareReq) (*pb.ActShareReply, error) {
	rly, e := s.actSvr.ActShare(ctx, req)
	if e != nil {
		return nil, e
	}
	if rly == nil {
		return &pb.ActShareReply{}, nil
	}
	return rly, nil
}

func (s *server) IndexSVideo(c context.Context, arg *pb.IndexSVideoReq) (*pb.IndexSVideoReply, error) {
	return s.svr.FeedIndexSvideo(c, arg.EntranceId, arg.Index)
}

func (s *server) AggrSVideo(c context.Context, arg *pb.AggrSVideoReq) (*pb.AggrSVideoReply, error) {
	return s.svr.AggrSvideo(c, arg.HotwordId, arg.Index)
}

func (s *server) SelectedSerie(c context.Context, req *pb.SelectedSerieReq) (*pb.SelectedSerieRly, error) {
	return s.svr.SelectedSerie(c, req.Type, req.Number)
}

func (s *server) BatchSerie(c context.Context, req *pb.BatchSerieReq) (*pb.BatchSerieRly, error) {
	return s.svr.BatchSerie(c, req.Type, req.Number)
}
