package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/ugc-season/job/conf"
	"go-gateway/app/app-svr/ugc-season/job/service"
)

var (
	ugcSeasonJob *service.Service
)

// Init init http router.
func Init(c *conf.Config, s *service.Service) {
	ugcSeasonJob = s
	e := bm.DefaultServer(c.BM)
	innerRouter(e)
	// init internal server
	if err := e.Start(); err != nil {
		log.Error("http.Serve error(%v)", err)
		panic(err)
	}
}

// innerRouter init inner router.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Any("railgun/coin", ugcSeasonJob.CoinSnRailgunHttp())
}

// ping check server ok.
func ping(c *bm.Context) {}
