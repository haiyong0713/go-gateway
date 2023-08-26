package http

import (
	"go-common/library/conf/paladin"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/service"
	arcmid "go-gateway/app/app-svr/archive/middleware"
)

var (
	svc     *service.Service
	authSvc *auth.Auth
)

// New new a bm server.
func New(s *service.Service) (*bm.Engine, error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err := paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return nil, err
	}
	if err := ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	svc = s
	authSvc = auth.New(nil)
	engine := bm.DefaultServer(&cfg)
	initRouter(engine)
	if err := engine.Start(); err != nil {
		return nil, err
	}
	return engine, nil
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/v2/feed-ng")
	{
		g.POST("/validate/session", validateSession)
		g.POST("/compare/session", compareSession)
		g.GET("/compare", comparePage)
		g.POST("/compare", compareFormSession)
		g.GET("/index", authSvc.GuestMobile, arcmid.BatchPlayArgs(), index)
	}
}

func ping(ctx *bm.Context) {}
