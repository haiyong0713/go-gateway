package grpc

import (
	pb "go-gateway/app/web-svr/datasource-ng/admin/api"

	"go-common/library/conf/paladin"
	"go-common/library/net/rpc/warden"
)

// New new a grpc server.
func New(svc pb.DatasourceServer) (ws *warden.Server, err error) {
	var rc struct {
		Server *warden.ServerConfig
	}
	err = paladin.Get("grpc.toml").UnmarshalTOML(&rc)
	if err == paladin.ErrNotExist {
		err = nil
	}
	ws = warden.NewServer(rc.Server)
	pb.RegisterDatasourceServer(ws.Server(), svc)
	ws, err = ws.Start()
	return
}
