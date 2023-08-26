package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/archive/job/conf"
	"go-gateway/app/app-svr/archive/job/service"
)

var (
	arcJobSrv *service.Service
)

// Init init http router.
func Init(c *conf.Config, s *service.Service) {
	arcJobSrv = s
	e := bm.DefaultServer(c.BM)
	innerRouter(e)
	// init internal server
	if err := e.Start(); err != nil {
		log.Error("xhttp.Serve error(%v)", err)
		panic(err)
	}
}

// innerRouter init inner router.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
	archiveJob := e.Group("/x/internal/v2/archive-job", bm.CSRF())
	{
		archiveJob.POST("/arc/update", arcUpdate)
	}
}

// ping check server ok.
func ping(c *bm.Context) {}
