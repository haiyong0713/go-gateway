package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/resource/service/conf"
	"go-gateway/app/app-svr/resource/service/service"
)

var (
	vfySvc *verify.Verify
	resSvc *service.Service
)

// Init int http service
func Init(c *conf.Config, s *service.Service) {
	vfySvc = verify.New(c.Verify)
	resSvc = s
	// init internal router
	engineInner := bm.DefaultServer(c.BM.Inner)
	innerRouter(engineInner)
	// init internal server
	if err := engineInner.Start(); err != nil {
		log.Error("engineInner.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
	// init external router
	engineLocal := bm.DefaultServer(c.BM.Local)
	localRouter(engineLocal)
	// init external server
	if err := engineLocal.Start(); err != nil {
		log.Error("engineLocal.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
}

// innerRouter init outer router api path.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Register(register)

	rs := e.Group("/x/internal/resource")

	resolution := rs.Group("/resolution")
	{
		resolution.GET("/limit/free", vfySvc.Verify, limitFree)
	}

	bn := rs.Group("/banner")
	bn.GET("", banner)

	ads := rs.Group("/ads")
	ads.GET("/paster/app", vfySvc.Verify, pasterAPP)
	ads.GET("/paster/pgc", vfySvc.Verify, pasterPGC)

	res := rs.Group("/res")
	res.GET("/indexIcon", vfySvc.Verify, indexIcon)
	res.GET("/playerIcon", vfySvc.Verify, playerIcon)
	res.GET("/playerPgcIcon", vfySvc.Verify, playerPgcIcon)
	res.GET("/cmtbox", vfySvc.Verify, cmtbox)
	res.GET("/regionCard", vfySvc.Verify, regionCard)
	res.GET("/audit", vfySvc.Verify, audit)
	res.GET("/customConfig", vfySvc.Verify, customConfig)
	res.GET("/isUploader", isUploader)
	res.GET("/isNotUploader", isNotUploader)

	dy := rs.Group("/dynamic")
	dy.GET("/search", vfySvc.Verify, dySearch)
}

// localRouter init local router api path.
func localRouter(e *bm.Engine) {
	e.GET("/x/resource/version", version)
	e.GET("/x/resource/monitor", monitor)
}
