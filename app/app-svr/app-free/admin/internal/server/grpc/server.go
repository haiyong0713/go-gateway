package grpc

import (
	"go-common/library/conf/paladin"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-free/admin/internal/service"
)

// New new a grpc server.
func New(svc *service.Service) *warden.Server {
	var rc struct {
		Server *warden.ServerConfig
	}
	if err := paladin.Get("grpc.toml").UnmarshalTOML(&rc); err != nil {
		if err != paladin.ErrNotExist {
			panic(err)
		}
	}
	ws := warden.NewServer(rc.Server)
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}
