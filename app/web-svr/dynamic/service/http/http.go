package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/dynamic/service/conf"
	"go-gateway/app/web-svr/dynamic/service/service"
)

var (
	dySvc  *service.Service
	vfySvr *verify.Verify
)

// Init init.
func Init(c *conf.Config, s *service.Service) {
	vfySvr = verify.New(c.Verify)
	dySvc = s
	engineInner := bm.DefaultServer(c.BM.Inner)
	innerRouter(engineInner)
	// init inner serve
	if err := engineInner.Start(); err != nil {
		log.Error("engineInner.Start error(%v)", err)
		panic(err)
	}
	// init external router
	enlocal := bm.DefaultServer(c.BM.Local)
	localRouter(enlocal)
	// init external server
	if err := enlocal.Start(); err != nil {
		log.Error("xhttp.Serve error(%v)", err)
		panic(err)
	}
}

// innerRouter init inner router.
func innerRouter(e *bm.Engine) {
	e.Use(bm.CORS())
	e.Ping(ping)
	e.Register(register)
	group := e.Group("/x/dynamic")
	{
		group.GET("/tag", vfySvr.Verify, regionTagArcs)
		group.GET("/region", vfySvr.Verify, regionArcs)
		group.GET("/regions", vfySvr.Verify, regionsArcs)
		group.GET("/region/total", vfySvr.Verify, regionTotal)
	}
}

// ping check server ok.
func ping(c *bm.Context) {
	if err := dySvc.Ping(c); err != nil {
		log.Error("dynamic service ping error")
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(nil, nil)
}

// localRouter init local router.
func localRouter(e *bm.Engine) {
	e.GET("/dynamic-service/redis/init/arc", initRegionArc)
}
