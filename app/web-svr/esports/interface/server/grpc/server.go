// Package grpc generate by warden_gen
package grpc

import (
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"

	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/service"
)

// New Resource warden rpc server
func New(c *warden.ServerConfig, svr *service.Service, quotaCfg *quota.Config) *warden.Server {
	limiter := quota.New(quotaCfg)
	ws := warden.NewServer(c)
	ws.Use(limiter.Limit())
	pb.RegisterEsportsServer(ws.Server(), svr)
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}
