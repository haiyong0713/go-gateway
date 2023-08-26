package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/dance-taiko/job/conf"
	"go-gateway/app/web-svr/dance-taiko/job/service"
)

var (
	srvweb *service.Service
)

// Init init
func Init(c *conf.Config, s *service.Service) {
	srvweb = s
	engine := bm.DefaultServer(c.BM)
	router(engine)
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

func router(e *bm.Engine) {
	e.Ping(ping)
	e.Register(register)
}

func ping(c *bm.Context) {
	if err := srvweb.Ping(c); err != nil {
		log.Error("dance-taiko-job ping error(%v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)
}
