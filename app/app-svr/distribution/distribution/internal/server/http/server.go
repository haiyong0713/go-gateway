package http

import (
	"go-common/library/conf/paladin.v2"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/distribution/distribution/internal/service"
)

//nolint:unused
var svc *service.Service

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
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
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	// g := e.Group("/distribution")
	// {
	// }
}

func ping(ctx *bm.Context) {}
