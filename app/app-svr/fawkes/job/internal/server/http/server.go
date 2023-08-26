package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-common/library/conf/paladin.v2"

	"github.com/google/wire"

	packSvc "go-gateway/app/app-svr/fawkes/job/internal/service/pack"
)

var Provider = wire.NewSet(New)

func New(pSvc *packSvc.Service) (engine *bm.Engine, err error) {
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
	engine = bm.DefaultServer(&cfg)
	initRouter(engine, pSvc)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine, s *packSvc.Service) {
	e.Ping(ping)
	job := e.Group("/x/job")
	r := job.Group("/railgun")
	{
		r.POST("/clearpack", s.CleanHttp)
	}
}

func ping(ctx *bm.Context) {

}
