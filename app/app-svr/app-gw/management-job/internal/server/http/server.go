package http

import (
	"go-common/library/conf/paladin.v2"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management-job/api"
)

var svc pb.ManagementJobServer

// New new a bm server.
func New(s pb.ManagementJobServer) (engine *bm.Engine, err error) {
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
	pb.RegisterManagementJobBMServer(engine, s)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	e.POST("/task/do", taskDo)
	e.GET("/config/raw", rawConfig)
}

func ping(ctx *bm.Context) {}
