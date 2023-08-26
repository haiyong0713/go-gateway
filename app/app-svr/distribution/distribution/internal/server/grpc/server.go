package grpc

import (
	pb "go-gateway/app/app-svr/distribution/distribution/api"
	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"

	authn "go-common/component/auth/middleware/grpc"
	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
)

// New new a grpc server.
func New(svc pb.DistributionServer) (ws *warden.Server, err error) {
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

	authN := authn.New(nil)

	ws = warden.NewServer(&cfg)
	ws.Use(authN.UnaryServerInterceptor(true), sessioncontext.UnaryServerInterceptor())

	pb.RegisterDistributionServer(ws.Server(), svc)
	ws, err = ws.Start()
	return
}
