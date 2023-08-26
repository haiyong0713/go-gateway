// Package server generate by warden_gen
package grpc

import (
	"go-common/library/net/rpc/warden"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
	"go-gateway/app/app-svr/resource/service/service"
)

// New Resource warden rpc server
func New(c *warden.ServerConfig, svr *service.Service) *warden.Server {
	ws := warden.NewServer(c)
	pb.RegisterResourceServer(ws.Server(), svr)
	pb2.RegisterResourceServer(ws.Server(), svr)
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}
