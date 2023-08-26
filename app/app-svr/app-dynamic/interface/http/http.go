package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	drawsvc "go-gateway/app/app-svr/app-dynamic/interface/service/draw"
	dynamicsvc "go-gateway/app/app-svr/app-dynamic/interface/service/dynamic"
	dynamicsvcV2 "go-gateway/app/app-svr/app-dynamic/interface/service/dynamicV2"
	topicsvc "go-gateway/app/app-svr/app-dynamic/interface/service/topic"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	dynamicSvc *dynamicsvc.Service
	verifySvc  *verify.Verify
	authSvc    *auth.Auth
	drawSvc    *drawsvc.Service
	topicSvc   *topicsvc.Service
)

type Server struct {
	DynamicSvc   *dynamicsvc.Service
	VerifySvc    *verify.Verify
	AuthSvc      *auth.Auth
	DrawSvc      *drawsvc.Service
	DynamicSvcV2 *dynamicsvcV2.Service
	Config       *conf.Config
	TopicSvc     *topicsvc.Service
	FeatureSvc   *feature.Feature
}

func Init(c *conf.Config, svr *Server) {
	initService(svr)
	// init external router
	engineOut := bm.DefaultServer(c.BM.Outer)
	outerRouter(engineOut)
	// init Outer server
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
}

// initService init services.
func initService(svr *Server) {
	dynamicSvc = svr.DynamicSvc
	verifySvc = svr.VerifySvc
	authSvc = svr.AuthSvc
	drawSvc = svr.DrawSvc
	topicSvc = svr.TopicSvc
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(bm.CORS(), anticrawler.Report())
	dynamic := e.Group("/x/bplus")
	{
		dynamic.GET("/feed/info", verifySvc.Verify, authSvc.GuestMobile, feedInfo)
		g := dynamic.Group("/draw/search", verifySvc.Verify)
		{
			g.GET("/all", drawImgTagRPCSearchAll)
			g.GET("/users", drawImgTagRPCSearchUsers)
			g.GET("/topics", drawImgTagRPCSearchTopics)
			g.GET("/locations", drawImgTagRPCSearchLocations)
			g.GET("/items", drawImgTagRPCSearchItems)
		}
		topic := dynamic.Group("/topic")
		{
			topic.GET("/square", authSvc.User, square)
			topic.GET("/hot/list", authSvc.User, hotList)
			topic.POST("/subscribe/save", authSvc.User, subscribeSave)
		}
		// 位置定位
		dynamic.GET("/geo", geoCoder)
	}
}
