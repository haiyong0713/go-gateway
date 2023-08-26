package grpc

import (
	pb "go-gateway/app/app-svr/app-gw/management/api"

	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
)

// New new a grpc server.
func New(svc pb.ManagementServer) (ws *warden.Server, err error) {
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
	ws = warden.NewServer(&cfg)
	pb.RegisterManagementServer(ws.Server(), svc)
	ws, err = ws.Start()
	return
}
