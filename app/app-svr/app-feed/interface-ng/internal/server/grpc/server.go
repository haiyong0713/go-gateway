package grpc

import (
	pb "go-gateway/app/app-svr/app-feed/interface-ng/api"

	"go-common/library/conf/paladin"
	"go-common/library/net/rpc/warden"
)

// New new a grpc server.
func New(svc pb.AppFeedNGServer) (*warden.Server, error) {
	var (
		cfg warden.ServerConfig
		ct  paladin.TOML
	)
	if err := paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return nil, err
	}
	if err := ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	ws := warden.NewServer(&cfg)
	pb.RegisterAppFeedNGServer(ws.Server(), svc)
	ws, err := ws.Start()
	if err != nil {
		return nil, err
	}
	return ws, nil
}
