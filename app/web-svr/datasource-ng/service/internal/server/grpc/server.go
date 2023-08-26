package grpc

import (
	"go-common/library/conf/paladin"
	"go-common/library/net/rpc/warden"
	pb "go-gateway/app/web-svr/datasource-ng/service/api"
)

// New new a grpc server.
func New(svc pb.DataSourceNGServer) (ws *warden.Server, err error) {
	var rc struct {
		Server *warden.ServerConfig
	}
	err = paladin.Get("grpc.toml").UnmarshalTOML(&rc)
	if err == paladin.ErrNotExist {
		err = nil
	}
	ws = warden.NewServer(rc.Server)
	pb.RegisterDataSourceNGServer(ws.Server(), svc)
	ws, err = ws.Start()
	return
}
