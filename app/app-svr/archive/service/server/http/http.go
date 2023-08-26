package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/archive/service/conf"
	"go-gateway/app/app-svr/archive/service/service"
)

var (
	idfSvc *verify.Verify
	arcSvc *service.Service
)

// Init init http router.
func Init(c *conf.Config, s *service.Service) {
	arcSvc = s
	idfSvc = verify.New(nil)
	// init internal router
	en := bm.DefaultServer(c.BM.Inner)
	innerRouter(en)
	// init internal server
	if err := en.Start(); err != nil {
		log.Error("xhttp.Serve error(%v)", err)
		panic(err)
	}
}

// innerRouter init inner router.
func innerRouter(e *bm.Engine) {
	e.Use(bm.CORS())
	e.Ping(ping)
	e.Register(register)
	archive := e.Group("/x/internal/v2/archive", bm.CSRF())
	{
		archive.GET("", idfSvc.Verify, arcInfo)
		archive.GET("/view", idfSvc.Verify, arcView)
		archive.GET("/views", idfSvc.Verify, arcViews)
		archive.GET("/page", idfSvc.Verify, arcPage)
		archive.GET("/video", idfSvc.Verify, video)
		archive.GET("/archives", idfSvc.Verify, archives)
		archive.GET("/archives/playurl", idfSvc.Verify, archivesWithPlayer)
		archive.GET("/typelist", idfSvc.Verify, typelist)
		archive.GET("/description", idfSvc.Verify, description)
		statGp := archive.Group("/stat")
		{
			statGp.GET("", idfSvc.Verify, arcStat)
			statGp.GET("/stats", idfSvc.Verify, arcStats)
		}
		upGp := archive.Group("/up")
		{
			upGp.GET("/count", idfSvc.Verify, upperCount)
			upGp.GET("/passed", idfSvc.Verify, upperPassed)
		}
	}
}

// ping check server ok.
func ping(c *bm.Context) {
}

func register(c *bm.Context) {
	c.JSON(nil, nil)
}
