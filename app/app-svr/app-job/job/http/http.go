package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-job/job/conf"
	"go-gateway/app/app-svr/app-job/job/service"
	"go-gateway/app/app-svr/app-job/job/service/dynamic"
	"go-gateway/app/app-svr/app-job/job/service/fawkes"
	"go-gateway/app/app-svr/app-job/job/service/feed"
	"go-gateway/app/app-svr/app-job/job/service/rank"
	ranklist "go-gateway/app/app-svr/app-job/job/service/rank-list"
	"go-gateway/app/app-svr/app-job/job/service/region"
	"go-gateway/app/app-svr/app-job/job/service/show"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	Svc         *service.Service
	ShowSvc     *show.Service
	FeedSvc     *feed.Service
	FawkesSvc   *fawkes.Service
	RegionSvc   *region.Service
	RankSvc     *rank.Service
	RankListSvc *ranklist.Service
	Feature     *feature.Feature
	DynamicSvc  *dynamic.Service
)

// Init init http
func Init(c *conf.Config) {
	Feature = feature.New(nil)
	initService(c)
	// init external router
	engineIn := bm.DefaultServer(c.BM.Inner)
	innerRouter(engineIn)
	// init Inner server
	if err := engineIn.Start(); err != nil {
		log.Error("bm.DefaultServer error(%v)", err)
		panic(err)
	}
}

func initService(c *conf.Config) {
	Svc = service.New(c)
	ShowSvc = show.New(c)
	FeedSvc = feed.New(c)
	FawkesSvc = fawkes.New(c)
	RegionSvc = region.New(c)
	RankSvc = rank.New(c)
	RankListSvc = ranklist.New(c)
	DynamicSvc = dynamic.New(c)
}

// innerRouter init inner router api path.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
}

func ping(c *bm.Context) {
	err := Svc.Ping(c)
	if err == nil {
		err = ShowSvc.Ping(c)
	}
	if err != nil {
		log.Error("app-job service ping error(%+v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
