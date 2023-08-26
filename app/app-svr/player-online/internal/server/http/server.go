package http

import (
	bm "go-common/library/net/http/blademaster"
	"net/http"

	"go-common/library/log"

	"go-gateway/app/app-svr/player-online/internal/conf"
	"go-gateway/app/app-svr/player-online/internal/service/online"
)

var (
	svc *online.Service
)

// Init init
func Init(c *conf.Config, s *online.Service) {
	svc = s
	engine := bm.DefaultServer(c.BM)
	route(engine)
	if err := engine.Start(); err != nil {
		log.Error("bm Start error(%v)", err)
		panic(err)
	}
}

func route(e *bm.Engine) {
	e.Ping(ping)
	e.Register(register)
}

func ping(ctx *bm.Context) {
	if err := svc.Ping(ctx); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)
}
