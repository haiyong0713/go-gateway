package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"

	"go-gateway/app/app-svr/app-channel/interface/conf"
	channelSvr "go-gateway/app/app-svr/app-channel/interface/service/channel"
	channelSvrV2 "go-gateway/app/app-svr/app-channel/interface/service/channel_v2"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	// depend service
	channelSvc   *channelSvr.Service
	verifySvc    *verify.Verify
	authSvc      *auth.Auth
	channelSvcV2 *channelSvrV2.Service
	featureSvc   *feature.Feature
)

func Init(c *conf.Config) {
	initService(c)
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
func initService(c *conf.Config) {
	channelSvc = channelSvr.New(c)
	verifySvc = verify.New(nil)
	authSvc = auth.New(nil)
	channelSvcV2 = channelSvrV2.New(c)
	featureSvc = feature.New(nil)
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(bm.CORS())
	cl := e.Group("/x/channel", verifySvc.Verify, featureSvc.BuildLimitHttp())
	{
		feed := cl.Group("/feed", authSvc.GuestMobile)
		{
			feed.GET("", arcmid.BatchPlayArgs(), index)
			feed.GET("/index", arcmid.BatchPlayArgs(), index2)
			feed.GET("/tab", tab)
			feed.GET("/tab/list", tablist)
		}
		cl.POST("/add", authSvc.UserMobile, subscribeAdd)
		cl.POST("/cancel", authSvc.UserMobile, subscribeCancel)
		cl.POST("/update", authSvc.UserMobile, subscribeUpdate)
		cl.GET("/list", authSvc.GuestMobile, list)
		cl.GET("/subscribe", authSvc.UserMobile, subscribe)
		cl.GET("/discover", authSvc.GuestMobile, discover)
		cl.GET("/category", authSvc.GuestMobile, category)
		cl.GET("/square", authSvc.GuestMobile, arcmid.BatchPlayArgs(), square)
		cl.GET("/mysub", authSvc.UserMobile, mysub)
	}
	clv2 := e.Group("/x/v2/channel", featureSvc.BuildLimitHttp())
	{
		clv2.GET("/tab", verifySvc.Verify, authSvc.GuestMobile, tab2)
		clv2.GET("/tab3", verifySvc.Verify, authSvc.GuestMobile, tab3)
		clv2.GET("/list", verifySvc.Verify, authSvc.GuestMobile, list2)
		clv2.GET("/mine", verifySvc.Verify, authSvc.GuestMobile, mine)
		clv2.POST("/sort", verifySvc.Verify, authSvc.UserMobile, channelSort)
		clv2.GET("/square", verifySvc.Verify, authSvc.GuestMobile, arcmid.BatchPlayArgs(), square2)
		clv2.GET("/square/alpha", verifySvc.Verify, authSvc.GuestMobile, arcmid.BatchPlayArgs(), squareAlpha)
		clv2.GET("/detail", verifySvc.Verify, authSvc.GuestMobile, detail)
		clv2.GET("/multiple", verifySvc.Verify, authSvc.GuestMobile, arcmid.BatchPlayArgs(), multiple)
		clv2.GET("/selected", verifySvc.Verify, authSvc.GuestMobile, arcmid.BatchPlayArgs(), selected)
		clv2.GET("/rank", rankList)
		clv2.GET("/share", verifySvc.Verify, share)
		clv2.GET("/red", verifySvc.Verify, authSvc.GuestMobile, red)
		clv2.GET("/recommend", verifySvc.Verify, authSvc.GuestMobile, arcmid.BatchPlayArgs(), channelRcmd)
		clv2.GET("/region/list", verifySvc.Verify, authSvc.GuestMobile, regionList)
		clv2.GET("/square2", verifySvc.Verify, authSvc.GuestMobile, arcmid.BatchPlayArgs(), square3)
		clv2.GET("/recommend2", verifySvc.Verify, authSvc.GuestMobile, arcmid.BatchPlayArgs(), channelRcmd2)
		clv2.GET("/home", authSvc.GuestMobile, arcmid.BatchPlayArgs(), home)
		// 新版本频道话题中心使用home2入口，调整返回的数据结构
		clv2.GET("/home2", authSvc.GuestMobile, arcmid.BatchPlayArgs(), home2)
		baike := clv2.Group("/baike", authSvc.GuestMobile)
		{
			baike.GET("/nav", baikeNav)
			baike.GET("/feed", arcmid.BatchPlayArgs(), baikeFeed)
		}
	}
}
