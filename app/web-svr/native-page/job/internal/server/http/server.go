package http

import (
	"go-common/library/railgun.v2"
	"net/http"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/web-svr/native-page/job/api"
)

var svc pb.NativePageJobServer

// New new a bm server.
func New(s pb.NativePageJobServer) (engine *bm.Engine, err error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svc = s
	engine = bm.DefaultServer(&cfg)
	pb.RegisterNativePageJobBMServer(engine, s)
	initRouter(engine)
	railgun.RegisterBMHandler(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
