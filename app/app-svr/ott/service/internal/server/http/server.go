package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/ott/service/conf"
	"go-gateway/app/app-svr/ott/service/internal/service"
)

var (
	svc *service.Service
)

// New new a bm server.
func New(c *conf.Config, s *service.Service) (engine *bm.Engine) {
	svc = s
	engine = bm.DefaultServer(c.Server)
	initRouter(engine)
	if err := engine.Start(); err != nil {
		panic(err)
	}
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
