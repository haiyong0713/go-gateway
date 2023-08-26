package grpc

import (
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/service"
)

// New new a grpc server.
func New(c *warden.ServerConfig, svc *service.Service, quotaCfg *quota.Config) (ws *warden.Server, err error) {
	limiter := quota.New(quotaCfg)
	ws = warden.NewServer(c)
	ws.Use(limiter.Limit())
	pb.RegisterEsportsServiceServer(ws.Server(), svc)
	ws, err = ws.Start()
	return
}
