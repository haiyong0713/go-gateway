package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/kvo/interface/conf"
	"go-gateway/app/app-svr/kvo/interface/service"
)

var (
	kvoSvr  *service.Service
	authSvr *auth.Auth
	vfySvr  *verify.Verify
)

// Init init http
func Init(c *conf.Config, svr *service.Service) {
	kvoSvr = svr

	authSvr = auth.New(c.Auth)
	vfySvr = verify.New(c.Verify)
	// init outer router
	engine := bm.DefaultServer(c.BM)
	outerRouter(engine)
	internalRouter(engine)
	if err := engine.Start(); err != nil {
		log.Error("engineOut.Start error(%v)", err)
		panic(err)
	}
}

func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(bm.CORS())
	group := e.Group("/x/kvo")
	{
		group.GET("/web/doc/get", authSvr.UserWeb, doc)
		group.POST("/web/doc/add", authSvr.UserWeb, addDoc)
		group.GET("/app/doc/get", authSvr.UserMobile, doc)
		group.POST("/app/doc/add", authSvr.UserMobile, addDoc)
	}
}

func internalRouter(e *bm.Engine) {
	group := e.Group("/x/internal/kvo")
	{
		group.GET("/doc/get", userConf)
	}
}

// ping check server ok.
func ping(c *bm.Context) {
	var err error
	if err = kvoSvr.Ping(c); err != nil {
		log.Error("kvo service ping error(%v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
