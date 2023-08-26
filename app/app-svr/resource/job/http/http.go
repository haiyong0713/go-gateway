package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/resource/job/conf"
	"go-gateway/app/app-svr/resource/job/service"
)

var (
	Svc *service.Service
)

// Init init http
func Init(c *conf.Config) {
	initService(c)
	// init external router
	engineIn := bm.DefaultServer(c.HttpService.Inner)
	innerRouter(engineIn)
	// init Inner server
	if err := engineIn.Start(); err != nil {
		log.Error("bm.DefaultServer error(%v)", err)
		panic(err)
	}
}

func initService(c *conf.Config) {
	Svc = service.New(c)
}

// innerRouter init inner router api path.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
}
