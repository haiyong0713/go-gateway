package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-wall/job/conf"
	"go-gateway/app/app-svr/app-wall/job/service/unicom"
)

var (
	unicomSvc *unicom.Service
)

// Init init
func Init(c *conf.Config) {
	initService(c)
	// init router
	engineInner := bm.DefaultServer(c.BM.Inner)
	outerRouter(engineInner)
	if err := engineInner.Start(); err != nil {
		log.Error("bm.DefaultServer error(%v)", err)
		panic(err)
	}
}

// initService init services.
func initService(c *conf.Config) {
	unicomSvc = unicom.New(c)
}

// Close
func Close() {
	unicomSvc.Close()
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	//init api
	e.Ping(ping)
}

// ping check server ok.
func ping(c *bm.Context) {
}
