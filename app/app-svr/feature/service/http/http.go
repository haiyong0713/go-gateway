package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/feature/service/conf"
	"go-gateway/app/app-svr/feature/service/service"
)

var featureSvr *service.Service

// Init int http service
func Init(c *conf.Config, s *service.Service) {
	featureSvr = s
	// init internal router
	engineInner := bm.DefaultServer(c.HTTPServers.Inner)
	innerRouter(engineInner)
	// init internal server
	if err := engineInner.Start(); err != nil {
		log.Error("engineInner.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
}

// innerRouter init outer router api path.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
	rs := e.Group("/x/internal/feature")

	bl := rs.Group("/build_limit")
	bl.GET("", buildLimit)
}
