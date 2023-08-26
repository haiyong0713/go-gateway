package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/proxy"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	commonSvr "go-gateway/app/app-svr/app-car/interface/service/common"
	"go-gateway/app/app-svr/app-car/interface/service/playurl"
	"go-gateway/app/app-svr/app-car/interface/service/reply"
	"go-gateway/app/app-svr/app-car/interface/service/resource"
	"go-gateway/app/app-svr/app-car/interface/service/show"
	"go-gateway/app/app-svr/app-car/interface/service/view"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	// depend service
	authSvc    *auth.Auth
	showSvc    *show.Service
	resSvc     *resource.Service
	viewSvc    *view.Service
	playSvc    *playurl.Service
	featureSvc *feature.Feature
	replySvc   *reply.Service
	verifySvc  *verify.Verify
	commonSvc  *commonSvr.Service
)

// Init is
func Init(c *conf.Config) {
	initService(c)
	// init external router
	engineOut := bm.DefaultServer(c.BM.Outer)
	outerRouter(engineOut, c)
	// init outer server
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start() error(%v)", err)
		panic(err)
	}
}

// initService init services.
func initService(c *conf.Config) {
	authSvc = auth.New(nil)
	showSvc = show.New(c)
	resSvc = resource.New(c)
	viewSvc = view.New(c)
	playSvc = playurl.New(c)
	featureSvc = feature.New(nil)
	replySvc = reply.New(c)
	verifySvc = verify.New(nil)
	commonSvc = commonSvr.New(c)
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine, _ *conf.Config) {
	e.Ping(ping)
	e.Use(bm.CORS())
	// formal api
	proxyHandler := proxy.NewZoneProxy("sh004", "http://sh001-app.bilibili.com")
	car := e.Group("/x/car", authSvc.GuestMobile)
	{
		s := car.Group("/show")
		{
			s.GET("", arcmid.BatchPlayArgs(), featureSvc.BuildLimitHttp(), proxyHandler, showIndex)
			s.GET("/tab", featureSvc.BuildLimitHttp(), showTab)
			s.GET("/pgc", showPGC)
			s.GET("/banner", arcmid.BatchPlayArgs(), proxyHandler, feedBanner)
			s.GET("/feed", arcmid.BatchPlayArgs(), proxyHandler, feedList)
		}
		car.GET("/banner", showBanner)
		car.GET("/mine", featureSvc.BuildLimitHttp(), mine)
		car.GET("/popular", popularIndex)
		sch := car.Group("/search")
		{
			sch.GET("", searchAll)
			sch.GET("/suggest", suggest)
		}
		car.GET("/dynamic/video", video)
		his := car.Group("/history")
		{
			his.GET("", historyList)
			his.POST("/report", proxyHandler, hisReport)
		}
		p := car.Group("/pgc")
		{
			p.GET("/myanime", myanime)
			p.GET("/list", bangumiList)
		}
		v := car.Group("/view")
		{
			v.GET("", viewIndex)
			v.GET("/relate", relateAll)
			v.POST("/like", proxyHandler, like)
			v.GET("/pgc/community", communityPGC)
		}
		car.GET("/space/archive", arcmid.BatchPlayArgs(), spaceArchive)
		fav := car.Group("/media")
		{
			fav.GET("/favorite", mediaFavorite)
			fav.GET("/list", mediaList)
			fav.GET("/topview", toview)
		}
		rg := car.Group("/region")
		{
			rg.GET("/list", regionList)
		}
		media := car.Group("/media")
		{
			media.GET("/popular", mediaPopular)
			//for小鹏美妆空间 近期热门 https://info.bilibili.co/pages/viewpage.action?pageId=346246042
			media.GET("/region", mediaRegion)
			//for小鹏美妆空间 定制的tab eg.清新裸妆
			media.GET("/search", mediaSearch)
		}
		car.GET("/fm/list", fmList)
		favFolder := car.Group("/fav")
		{
			favFolder.GET("/folder/list", userFolders)
			favFolder.POST("/folder/create", proxyHandler, addFolder)
			favFolder.POST("/folder/batchdea", proxyHandler, favAddOrDelFolders)
		}
		player := car.Group("/player")
		{
			//播放链接
			player.GET("/playurl", playurlApp)
		}
		//行为上报，for杜比
		car.GET("/event/report", eventReport)
		reply := car.Group("/reply")
		{
			reply.GET("", replyList)
			reply.GET("/list", replyChild)
		}
		receive := car.Group("/receive")
		{
			receive.POST("/vip", verifySvc.Verify, addVip)
			receive.POST("/vip/code", verifySvc.Verify, codeOpen)
		}
		audio := car.Group("/audio")
		{
			audio.GET("", audioShow)
			audio.GET("/feed", audioFeed)
			audio.GET("/channel", audioChannel)
			audio.POST("/report/play", reportPlayAction)
		}
	}
	carWeb := e.Group("/x/car/web", authSvc.GuestWeb, bm.CSRF())
	{
		s := carWeb.Group("/show")
		{
			s.GET("", arcmid.BatchPlayArgs(), proxyHandler, showWebIndex)
			s.GET("/tab", showTabWeb)
			s.GET("/pgc", showPGCWeb)
		}
		carWeb.GET("/mine", mineWeb)
		carWeb.GET("/popular", popularIndexWeb)
		sch := carWeb.Group("/search")
		{
			sch.GET("", searchWebAll)
			sch.GET("/suggest", suggestWeb)
		}
		carWeb.GET("/dynamic/video", videoWeb)
		his := carWeb.Group("/history")
		{
			his.GET("", historyWebList)
			his.POST("/report", proxyHandler, hisReportWeb)
		}
		p := carWeb.Group("/pgc")
		{
			p.GET("/myanime", myanimeWeb)
			p.GET("/list", bangumiListWeb)
		}
		v := carWeb.Group("/view")
		{
			v.GET("", ViewWeb)
			v.GET("/relate", relateWebAll)
			v.POST("/like", proxyHandler, likeWeb)
			v.GET("/pgc/community", communityWebPGC)
		}
		rg := carWeb.Group("/region")
		{
			rg.GET("/list", regionListWeb)
		}
		media := carWeb.Group("/media")
		{
			media.GET("/popular", mediaPopularWeb)
			media.GET("/search", mediaSearchWeb)
			media.GET("/favorite", favoriteWeb)
			media.GET("/list", mediaListWeb)
			media.GET("/topview", toviewWeb)
			media.GET("/pgc", mediaPGCWeb)
		}
		player := carWeb.Group("/player")
		{
			player.GET("/playurl", playurlWeb)
		}
		favFolder := carWeb.Group("/fav")
		{
			favFolder.POST("/folder/create", addFolderWeb)
			favFolder.POST("/folder/batchdea", favAddOrDelFoldersWeb)
			favFolder.GET("/folder/list", userFolders)
		}
		carWeb.GET("/space/archive", spaceArchiveWeb)
		reply := carWeb.Group("/reply")
		{
			reply.GET("", replyList)
			reply.GET("/list", replyChild)
		}
	}
	// 车载2.0接口
	carV2 := e.Group("/x/v2/car", authSvc.GuestMobile, arcmid.BatchPlayArgs())
	{
		// single

		carV2.GET("/mine/tabs", mineV2Tabs)
		carV2.GET("/dynamic", dynamicV2)
		carV2.GET("/search", searchV2)
		carV2.GET("/space", spaceV2)
		carV2.GET("/region_meta", regionMeta)
		// group
		continueV2 := carV2.Group("/continue")
		{
			continueV2.GET("", viewHistory)
			continueV2.GET("/tab", viewHistoryTab)
			continueV2.GET("/tab/more", viewHistoryTabMore)
		}
		viewV2 := carV2.Group("/view")
		{
			viewV2.GET("", viewV2Detail)
			viewV2.GET("/rcmd", viewV2Rcmd)
			viewV2.GET("/serial", viewV2Serial)
		}
		favV2 := carV2.Group("/favorite")
		{
			favV2.GET("", favoriteV2)
			favV2.GET("/video", favoriteVideoV2)
			favV2.GET("/bangumi", favoriteBangumiV2)
			favV2.GET("/cinema", favoriteCinemaV2)
			favV2.GET("/toview", favoriteToView)
		}
		payV2 := carV2.Group("/pay")
		{
			payV2.GET("/info", payInfo)
			payV2.GET("/result", payResult)
		}
		// fm
		fmV2 := carV2.Group("/fm")
		{
			fmV2.GET("/show", fmShow)
			fmV2.GET("/show_v2", fmShowV2)
			fmV2.GET("/pin_page", pinPage)
			fmV2.GET("/list", fmListRefactor)
			fmV2.POST("/like", fmLike)
		}
		videoV2 := carV2.Group("/video")
		{
			videoV2.GET("/tabs", videoTabs)
			videoV2.GET("/tab/cards", videoTabCards)
			videoV2.GET("/tab/cards/playlist", cardPlaylist)
		}
	}

	// 车载2.0 Web接口
	carWebV2 := e.Group("/x/v2/car/web", authSvc.GuestWeb, bm.CSRF(), arcmid.BatchPlayArgs())
	{
		// single
		carWebV2.GET("/mine/tabs", mineV2Tabs)
		carWebV2.GET("/dynamic", dynamicV2)
		carWebV2.GET("/search", searchV2)
		carWebV2.GET("/space", spaceV2)
		carWebV2.GET("/media/parse", mediaParse)
		// group
		webContinueV2 := carWebV2.Group("/continue")
		{
			webContinueV2.GET("", viewHistory)
			webContinueV2.GET("/tab", viewHistoryTabWeb)
			webContinueV2.GET("/tab/more", viewHistoryTabMore)
		}
		webViewV2 := carWebV2.Group("/view")
		{
			webViewV2.GET("", viewV2Detail)
			webViewV2.GET("/rcmd", viewV2Rcmd)
			webViewV2.GET("/serial", viewV2Serial)
		}
		webFavV2 := carWebV2.Group("/favorite")
		{
			webFavV2.GET("", favoriteV2)
			webFavV2.GET("/video", favoriteVideoV2)
			webFavV2.GET("/bangumi", favoriteBangumiV2)
			webFavV2.GET("/cinema", favoriteCinemaV2)
		}
		webPayV2 := carWebV2.Group("/pay")
		{
			webPayV2.GET("/info", payInfo)
			webPayV2.GET("/result", payResult)
		}
		// fm
		webFmV2 := carWebV2.Group("/fm")
		{
			webFmV2.GET("/show", fmShow)
			webFmV2.GET("/show_v2", fmShowV2)
			webFmV2.GET("/pin_page", pinPage)
			webFmV2.GET("/list", fmListRefactor)
			webFmV2.POST("/like", fmLike)
		}
		videoV2 := carWebV2.Group("/video")
		{
			videoV2.GET("/tabs", videoTabs)
			videoV2.GET("/tab/cards", videoTabCards)
			videoV2.GET("/tab/cards/playlist", cardPlaylist)
		}
	}
}

func pagePn(list []card.Handler, pn int) int {
	if len(list) > 0 {
		return pn + 1
	}
	return pn
}

func guestIdFromCtx(c *bm.Context) (int64, error) {
	guest, ok := c.Get("guest_id")
	if !ok {
		return 0, ecode.NoLogin
	}
	guestId, ok := guest.(int64)
	if !ok {
		return 0, ecode.NoLogin
	}
	return guestId, nil
}
