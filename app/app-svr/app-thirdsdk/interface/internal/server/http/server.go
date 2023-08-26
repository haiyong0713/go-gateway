package http

import (
	"go-common/library/conf/paladin.v2"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-thirdsdk/interface/internal/service"
)

var svc *service.Service

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
	var cfg bm.ServerConfig
	ac := &paladin.TOML{}
	if err = paladin.Watch("http.toml", ac); err != nil {
		return
	}
	if err = ac.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svc = s
	engine = bm.DefaultServer(&cfg)
	initRouter(engine, ac)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine, ac *paladin.Map) {
	e.Ping(ping)
	g := e.Group("/x/thirdsdk")
	{
		g.GET("/playurl", Verify(ac, true), playURL)
		g.POST("/user/binding/sync", Verify(ac, false), userBindSync)
		g.POST("/archive/status/sync", Verify(ac, false), arcStatusSync)
	}
	dm := g.Group("/dm")
	{
		dm.GET("/seg", Verify(ac, true), dmSeg)
	}
}

func ping(ctx *bm.Context) {}
