package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"

	pb "go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/conf"
	"go-gateway/app/app-svr/steins-gate/service/internal/service"
)

// New History warden rpc server
func New(cfg *conf.Config, c *warden.ServerConfig, svr *service.Service) *warden.Server {
	ws := warden.NewServer(c)
	ws.Use(quota.New(cfg.Quota).Limit())
	pb.RegisterSteinsGateServer(ws.Server(), &server{svr})
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}

type server struct {
	svr *service.Service
}

var _ pb.SteinsGateServer = &server{}

// RefreshSeriesList refreshes  the series list cache in redis
func (s *server) GraphInfo(ctx context.Context, req *pb.GraphInfoReq) (resp *pb.GraphInfoReply, err error) {
	var a *pb.GraphInfo
	resp = new(pb.GraphInfoReply)
	if a, err = s.svr.GraphInfo(ctx, req.Aid); err != nil {
		return
	}
	resp.GraphInfo = a
	return
}

func (s *server) View(ctx context.Context, req *pb.ViewReq) (resp *pb.ViewReply, err error) {
	return s.svr.View(ctx, req)
}

func (s *server) Views(ctx context.Context, req *pb.ViewsReq) (resp *pb.ViewsReply, err error) {
	return s.svr.Views(ctx, req)
}

func (s *server) GraphView(ctx context.Context, req *pb.GraphViewReq) (resp *pb.GraphViewReply, err error) {
	resp = new(pb.GraphViewReply)
	resp.Page, resp.Graph, resp.Evaluation, err = s.svr.GraphView(ctx, req.Aid)
	return
}

func (s *server) Evaluation(ctx context.Context, req *pb.EvaluationReq) (resp *pb.EvaluationReply, err error) {
	resp = new(pb.EvaluationReply)
	resp.Eval, err = s.svr.Evaluation(ctx, req.Aid)
	return
}

func (s *server) GraphRights(c context.Context, req *pb.GraphRightsReq) (resp *pb.GraphRightsReply, err error) {
	resp = new(pb.GraphRightsReply)
	resp.AllowPlay, err = s.svr.GraphRights(c, req)
	return
}

func (s *server) MarkEvaluations(c context.Context, req *pb.MarkEvaluationsReq) (resp *pb.MarkEvaluationsReply, err error) {
	resp = new(pb.MarkEvaluationsReply)
	resp.Items, err = s.svr.MarkEvaluations(c, req.Mid, req.Aids)
	return

}
