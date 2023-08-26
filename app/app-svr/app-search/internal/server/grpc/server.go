package grpc

import (
	mauth "go-common/component/auth/middleware/grpc"
	mrestrict "go-common/component/restriction/middleware/grpc"
	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"

	pb "go-gateway/app/app-svr/app-search/api/v1"
)

// New new a grpc server.
func New(svc pb.SearchServer) (ws *warden.Server, err error) {
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
	pb.RegisterSearchServer(ws.Server(), svc)
	// 用户鉴权
	auther := mauth.New(nil)
	ws.Add("/bilibili.app.interface.v1.Search/Suggest3", auther.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor())
	ws.Add("/bilibili.app.interface.v1.Search/DefaultWords", auther.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor())
	ws, err = ws.Start()
	return
}
