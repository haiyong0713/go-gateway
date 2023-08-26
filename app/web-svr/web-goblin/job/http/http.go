package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web-goblin/job/conf"
	"go-gateway/app/web-svr/web-goblin/job/service/web"
)

var (
	srvweb *web.Service
)

// Init init
func Init(c *conf.Config, s *web.Service) {
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
	e.GET("/load/change/out/arc", loadChangeOutArc)
}

func loadChangeOutArc(c *bm.Context) {
	srvweb.LoadChangeOutArc()
}

func ping(c *bm.Context) {
	if err := srvweb.Ping(c); err != nil {
		log.Error("web-goblin-job ping error(%v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)
}
