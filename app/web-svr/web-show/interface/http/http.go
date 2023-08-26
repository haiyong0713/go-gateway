package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/cache"
	"go-common/library/net/http/blademaster/middleware/cache/store"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/web-show/interface/conf"
	"go-gateway/app/web-svr/web-show/interface/service/job"
	"go-gateway/app/web-svr/web-show/interface/service/operation"
	ressvr "go-gateway/app/web-svr/web-show/interface/service/resource"
)

var (
	jobSvc  *job.Service
	opSvc   *operation.Service
	vfySvr  *verify.Verify
	authSvr *auth.Auth
	resSvc  *ressvr.Service

	// cache components
	cacheSvr *cache.Cache
	deg      *cache.Degrader
)

// Init int http service
func Init(c *conf.Config) {
	initService(c)
	// init external router
	cacheSvr = cache.New(store.NewMemcache(c.DegradeConfig.Memcache))
	deg = cache.NewDegrader(c.DegradeConfig.Expire)
	engineOut := bm.DefaultServer(c.BM.Outer)
	outerRouter(engineOut)
	// init Outer serve
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start error(%v)", err)
		panic(err)
	}
	engineLocal := bm.DefaultServer(c.BM.Local)
	localRouter(engineLocal)
	// init Outlocaler serve
	if err := engineLocal.Start(); err != nil {
		log.Error("engineLocal.Start error(%v)", err)
		panic(err)
	}
}

// initService init service
func initService(c *conf.Config) {
	authSvr = auth.New(c.Auth)
	vfySvr = verify.New(c.Verify)
	jobSvc = job.New(c)
	opSvc = operation.New(c)
	resSvc = ressvr.New(c)
}

// CloseService close all service
func CloseService() {
	jobSvc.Close()
	opSvc.Close()
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	// init api
	e.Use(bm.CORS())
	e.Ping(ping)
	group := e.Group("/x/web-show")
	{
		group.GET("/join", join)
		group.GET("/notice", notice)
		group.GET("/promote", promote)
		group.GET("/res/loc", authSvr.Guest, cacheSvr.Cache(deg.Args("id", "pf"), nil), resource)
		group.GET("/res/locs", authSvr.Guest, cacheSvr.Cache(deg.Args("ids", "pf"), nil), resources)
		group.GET("/ad/video", authSvr.Guest, advideo)
		group.GET("/archive/relation", relation)
		group.GET("/urls", vfySvr.Verify, urlMonitor)
		group.GET("/res/frontpage", frontPage)
		group.GET("/page/header", pageHeader)
	}
	e.GET("/x/ad/video", authSvr.Guest, advideo)
}

// innerRouter init local router api path.
func localRouter(e *bm.Engine) {
	e.Use(bm.CORS())
	group := e.Group("/x/web-show")
	{
		group.GET("/monitor/ping", ping)
		group.GET("/version", version)
		group.GET("/gray", grayRate)
	}
}
