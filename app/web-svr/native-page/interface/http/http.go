package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/http/blademaster/middleware/verify"

	"go-gateway/app/web-svr/native-page/interface/conf"
	"go-gateway/app/web-svr/native-page/interface/service/like"
)

var (
	likeSvc  *like.Service
	matchSvc *like.Service
	authSvc  *auth.Auth
	vfySvc   *verify.Verify
)

// Init int http service
func Init(c *conf.Config) {
	initService(c)
	engine := bm.NewServer(c.HTTPServer)
	limiter := quota.New(c.Limiter)
	engine.Use(bm.Recovery(), bm.Trace(), bm.Logger(), bm.Mobile(), bm.NewRateLimiter(nil).Limit(), limiter.Handler())
	outerRouter(engine)
	internalRouter(engine)
	// init Outer serve
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

// initService init service
func initService(c *conf.Config) {
	authSvc = auth.New(c.Auth)
	likeSvc = like.New(c)
	matchSvc = like.New(c)
	vfySvc = verify.New(c.Verify)
}

// CloseService close all service
func CloseService() {
	likeSvc.Close()
	matchSvc.Close()
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Use(bm.CORS())
	e.Ping(ping)
	e.Register(register)
	group := e.Group("/x/native_page", bm.CSRF())
	{
		dynamicGroup := group.Group("/dynamic")
		{
			dynamicGroup.GET("/index", authSvc.Guest, actIndex)
			dynamicGroup.GET("/menu", authSvc.Guest, menuTab)
			dynamicGroup.GET("/inline", authSvc.Guest, inlineTab)
			dynamicGroup.GET("/pages", natPages)
			dynamicGroup.GET("/topic", authSvc.Guest, actDynamic)
			dynamicGroup.GET("/new/video/aid", authSvc.Guest, newVideoAid)
			dynamicGroup.GET("/new/video/dyn", authSvc.Guest, newVideoDyn)
			dynamicGroup.GET("/resource/aid", authSvc.Guest, resourceAid)
			dynamicGroup.GET("/resource/dyn", authSvc.Guest, resourceDyn)
			dynamicGroup.GET("/season/ssid", authSvc.Guest, seasonIDs)
			dynamicGroup.GET("/season/source", authSvc.Guest, seasonSource)
			dynamicGroup.GET("/resource/role", authSvc.Guest, resourceRole)
			dynamicGroup.GET("/resource/origin", authSvc.Guest, resourceOrigin)
			dynamicGroup.GET("/editor/origin", authSvc.Guest, editorOrigin)
			dynamicGroup.GET("/editor/viewed_arcs", authSvc.User, edViewedArcs)
			dynamicGroup.GET("/timeline/source", timelineSource)
			dynamicGroup.GET("/module", natModule)
			dynamicGroup.GET("/live", authSvc.Guest, liveDyn)
			dynamicGroup.GET("/ts/mine/pages", authSvc.User, minePages)
			dynamicGroup.GET("/ts/act/pages", authSvc.User, upActPages)
			dynamicGroup.GET("/ts/page", authSvc.User, tsPage)
			dynamicGroup.POST("/ts/page/add", authSvc.User, minePageAdd)
			dynamicGroup.POST("/ts/page/save", authSvc.User, minePageSave)
			dynamicGroup.GET("/ts/remark", tsRemark)
			dynamicGroup.GET("/ts/white", authSvc.User, tsWhite)
			dynamicGroup.GET("/ts/setting", authSvc.User, tsSetting)
			dynamicGroup.GET("/ts/space", authSvc.User, tsSpace)
			dynamicGroup.POST("/ts/space/save", authSvc.User, tsSpaceSave)
			dynamicGroup.POST("/ts/white/save", authSvc.User, tsWhiteSave)
			dynamicGroup.GET("/archive/my", authSvc.User, myArchiveList)
			dynamicGroup.GET("/archive/activity", actArchiveList)
			dynamicGroup.GET("/progress", authSvc.Guest, progress)
			dynamicGroup.GET("/partition", authSvc.Guest, partition)
			dynamicGroup.GET("/partition/v2", authSvc.Guest, partitionV2)
		}
	}
}

func internalRouter(e *bm.Engine) {
	group := e.Group("/x/internal/native_page", bm.CSRF())
	{
		group.POST("/cache/clear", clearCache)
		group.GET("/dynamic/ts/white", vfySvc.Verify, inlineTsWhite)
	}
}

func ping(c *bm.Context) {
	if err := likeSvc.Ping(c); err != nil {
		log.Error("native-page interface ping error(%v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(nil, nil)
}
