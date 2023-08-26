package grpc

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"
	pb "go-gateway/app/app-svr/up-archive/service/api"

	"github.com/pkg/errors"
)

// New new a grpc server.
func New(svc pb.UpArchiveServer) (ws *warden.Server, err error) {
	var (
		ct       paladin.TOML
		cfg      warden.ServerConfig
		quotaCfg *quota.Config
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = ct.Get("quota").UnmarshalTOML(&quotaCfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	limiter := quota.New(quotaCfg)
	ws = warden.NewServer(&cfg)
	ws.Use(limiter.Limit())
	pb.RegisterUpArchiveServer(ws.Server(), svc)
	ws, err = ws.Start()
	return
}
