package grpc

import (
	"context"

	pb "go-gateway/app/app-svr/ott/service/api"

	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/ott/service/internal/service"
)

func New(c *warden.ServerConfig, svr *service.Service) *warden.Server {
	ws := warden.NewServer(c)
	pb.RegisterOTTServiceServer(ws.Server(), &server{s: svr})
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}

type server struct {
	s *service.Service
}

var _ pb.OTTServiceServer = &server{}

func (s *server) ArcsAllow(ctx context.Context, req *pb.ArcsAllowReq) (resp *pb.ArcsAllowReply, err error) {
	return s.s.ArcsAllow(ctx, req.Aids)
}
