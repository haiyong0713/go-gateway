package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/job/conf"
	"go-gateway/app/app-svr/app-car/job/service/duertv"
	"go-gateway/app/app-svr/app-car/job/service/fm"
)

var (
	duertvSvc *duertv.Service
	fmSvc     *fm.Service
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
	duertvSvc = duertv.New(c)
	fmSvc = fm.New(c)
}

// Close
func Close() {
	duertvSvc.Close()
	fmSvc.Close()
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	//init api
	e.Ping(ping)
}

// ping check server ok.
func ping(c *bm.Context) {
}
