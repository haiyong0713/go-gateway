package grpc

import (
	"context"

	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
	pb "go-gateway/app/app-svr/collection-splash/api"
	"go-gateway/app/app-svr/collection-splash/internal/service"

	"github.com/golang/protobuf/ptypes/empty"
)

type Server struct {
	svr *service.Service
}

// New new a grpc server.
func New(svc *service.Service) (ws *warden.Server, err error) {
	var (
		cfg warden.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	s := &Server{
		svr: svc,
	}
	ws = warden.NewServer(&cfg)
	pb.RegisterCollectionSplashServer(ws.Server(), s)
	ws, err = ws.Start()
	return
}

func (s *Server) AddSplash(ctx context.Context, arg *pb.AddSplashReq) (*pb.SetSplashReply, error) {
	return s.svr.AddSplash(ctx, arg)
}

func (s *Server) UpdateSplash(ctx context.Context, arg *pb.UpdateSplashReq) (*pb.SetSplashReply, error) {
	return s.svr.UpdateSplash(ctx, arg)
}

func (s *Server) DeleteSplash(ctx context.Context, arg *pb.SplashReq) (*pb.SetSplashReply, error) {
	return s.svr.DeleteSplash(ctx, arg)
}

func (s *Server) Splash(ctx context.Context, arg *pb.SplashReq) (*pb.SplashReply, error) {
	return s.svr.Splash(ctx, arg)
}

func (s *Server) SplashList(ctx context.Context, arg *empty.Empty) (*pb.SplashListReply, error) {
	return s.svr.SplashList(ctx, arg)
}
