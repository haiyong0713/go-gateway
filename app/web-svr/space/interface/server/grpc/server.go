// Package grpc generate by warden_gen
package grpc

import (
	"go-common/library/net/rpc/warden"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/service"
)

// New Resource warden rpc server
func New(c *warden.ServerConfig, svr *service.Service) *warden.Server {
	ws := warden.NewServer(c)
	pb.RegisterSpaceServer(ws.Server(), svr)
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}
