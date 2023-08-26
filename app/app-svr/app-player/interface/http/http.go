package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"

	"go-gateway/app/app-svr/app-player/interface/conf"
	"go-gateway/app/app-svr/app-player/interface/service"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	svr        *service.Service
	ver        *verify.Verify
	ah         *auth.Auth
	featuresvr *feature.Feature
)

// Init init http
func Init(c *conf.Config, featureSvr *feature.Feature) {
	featuresvr = featureSvr
	initService(c)
	engine := bm.DefaultServer(nil)
	outerRouter(engine)
	if err := engine.Start(); err != nil {
		panic(err)
	}
}

func initService(c *conf.Config) {
	svr = service.New(c)
	ver = verify.New(nil)
	ah = auth.New(nil)
}

func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(anticrawler.Report())
	player := e.Group("/x/playurl", ver.Verify)
	player.GET("", ah.GuestMobile, playurl)
	player.GET("/ott", ah.GuestMobile, playurlOtt)
	player.GET("/download/num", ah.UserMobile, dlNum)
	player.GET("/hls", ah.GuestMobile, playurlHls)
	player.GET("/hls/master.m3u8", ah.GuestMobile, featuresvr.BuildLimitHttp(), hlsMaster)
	player.GET("/hls/stream.m3u8", ah.GuestMobile, hlsStream)
	player.GET("/bubble", ah.UserMobile, bubble)
	player.POST("/bubble_submit", ah.UserMobile, bubbleSubmit)
	player.GET("/activity/proj_page", ah.GuestMobile, projPageAct)
	player.GET("/proj/activity", ah.GuestMobile, projActAll)
}

// Ping is
func ping(ctx *bm.Context) {

}
