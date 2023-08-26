package grpc

import (
	pb "go-gateway/app/web-svr/native-page/job/api"

	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
)

// New new a grpc server.
func New(svc pb.NativePageJobServer) (ws *warden.Server, err error) {
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
	pb.RegisterNativePageJobServer(ws.Server(), svc)
	ws, err = ws.Start()
	return
}
