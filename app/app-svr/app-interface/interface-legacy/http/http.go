package http

import (
	abtest "go-common/component/tinker/middleware/http"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/cache"
	"go-common/library/net/http/blademaster/middleware/cache/store"
	"go-common/library/net/http/blademaster/middleware/proxy"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/queue/databus"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	acc "go-gateway/app/app-svr/app-interface/interface-legacy/service/account"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/dataflow"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/display"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/favorite"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/history"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/media"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/relation"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/search"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/space"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/teenagers"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	verifySvc   *verify.Verify
	authSvc     *auth.Auth
	spaceSvr    *space.Service
	srcSvr      *search.Service
	displaySvr  *display.Service
	favSvr      *favorite.Service
	accSvr      *acc.Service
	relSvr      *relation.Service
	historySvr  *history.Service
	mediaSvr    *media.Service
	teenSvr     *teenagers.Service
	dataflowSvr *dataflow.Service
	// databus
	userActPub *databus.Databus
	config     *conf.Config
	// cache components
	cache2Svr *cache.Cache
	deg2      *cache.Degrader
	// feature service
	featureSvc *feature.Feature
)

type Server struct {
	VerifySvc   *verify.Verify
	AuthSvc     *auth.Auth
	SpaceSvr    *space.Service
	SrcSvr      *search.Service
	DisplaySvr  *display.Service
	FavSvr      *favorite.Service
	AccSvr      *acc.Service
	RelSvr      *relation.Service
	HistorySvr  *history.Service
	MediaSvr    *media.Service
	TeenSvr     *teenagers.Service
	DataflowSvr *dataflow.Service
	// databus
	UserActPub *databus.Databus
	Config     *conf.Config
	// cache components
	CacheSvr *cache.Cache
	Deg      *cache.Degrader
	// feature service
	FeatureSvc *feature.Feature
}

// Init init http
func Init(c *conf.Config, svr *Server) {
	initService(c, svr)
	// init external router
	cache2Svr = cache.New(store.NewMemcache(c.Degrade2Config.Memcache))
	deg2 = cache.NewDegrader(c.Degrade2Config.Expire)
	engineOut := bm.DefaultServer(c.BM.Outer)
	limiter := quota.New(c.QuotaConf)
	engineOut.Use(limiter.Handler())
	// init outer router
	outerRouter(engineOut)
	internalRouter(engineOut)
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
}

func initService(c *conf.Config, svr *Server) {
	verifySvc = svr.VerifySvc
	authSvc = svr.AuthSvc
	spaceSvr = svr.SpaceSvr
	srcSvr = svr.SrcSvr
	displaySvr = svr.DisplaySvr
	favSvr = svr.FavSvr
	accSvr = svr.AccSvr
	relSvr = svr.RelSvr
	historySvr = svr.HistorySvr
	mediaSvr = svr.MediaSvr
	teenSvr = svr.TeenSvr
	dataflowSvr = svr.DataflowSvr
	userActPub = svr.UserActPub
	config = c
	featureSvc = svr.FeatureSvc
}

func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(anticrawler.Report())
	proxyHandler := proxy.NewZoneProxy("sh004", "http://sh001-app.bilibili.com")

	space := e.Group("/x/v2/space", featureSvc.BuildLimitHttp())
	{
		space.GET("", verifySvc.Verify, authSvc.GuestMobile, arcmid.BatchPlayArgs(), spaceAll)
		upArchiveArgs := []string{"vmid", "pn", "ps", "mobi_app", "device", "build", "order", "clocale", "slocale", "s_locale", "c_locale"}
		space.GET("/archive", authSvc.GuestMobile, arcmid.BatchPlayArgs(), cache2Svr.Cache(deg2.Args(upArchiveArgs...), nil), upArchive)
		upCursorArchiveArgs := []string{"vmid", "aid", "from_view_aid", "ps", "mobi_app", "device", "sort", "order", "clocale", "slocale", "s_locale", "c_locale", "include_cursor"}
		space.GET("/archive/cursor", authSvc.GuestMobile, arcmid.BatchPlayArgs(), cache2Svr.Cache(deg2.Args(upCursorArchiveArgs...), nil), upArchiveCursor)
		space.GET("/series", authSvc.GuestMobile, arcmid.BatchPlayArgs(), upSeries)
		space.GET("/game", authSvc.GuestMobile, playGame)
		space.GET("/season", authSvc.GuestMobile, upSeason)
		space.GET("/season/videos", authSvc.GuestMobile, arcmid.BatchPlayArgs(), seasonView)
		space.GET("/comic", authSvc.GuestMobile, upComic)
		space.GET("/subcomic", authSvc.GuestMobile, subComic)
		space.GET("/article", authSvc.GuestMobile, upArticle)
		space.GET("/bangumi", authSvc.GuestMobile, bangumi)
		space.GET("/coinarc", authSvc.GuestMobile, coinArc)
		space.POST("/coin/cancel", proxyHandler, verifySvc.Verify, authSvc.UserMobile, coinCancel)
		space.GET("/likearc", authSvc.GuestMobile, likeArc)
		space.GET("/community", authSvc.GuestMobile, community)
		space.GET("/contribute", proxyHandler, authSvc.GuestMobile, contribute)
		space.GET("/contribute/cursor", proxyHandler, authSvc.GuestMobile, contribution)
		space.GET("/clips", authSvc.GuestMobile, clips)
		space.GET("/albums", authSvc.GuestMobile, albums)
		space.POST("/report", verifySvc.Verify, report)
		space.POST("/upContribute", proxyHandler, verifySvc.Verify, upContribute)
		space.GET("/upper/recmd", authSvc.GuestMobile, upperRecmd)
		searchArgs := []string{"keyword", "is_title", "highlight", "vmid", "pn", "ps"}
		space.GET("/search", authSvc.GuestMobile, cache2Svr.Cache(deg2.Args(searchArgs...), nil), spaceSearch)
		space.POST("/topphoto/reset", proxyHandler, authSvc.UserMobile, topphotoReset)
		space.POST("/attention/mark", verifySvc.Verify, authSvc.UserMobile, attentionMark)
		space.GET("/photo/top/list", authSvc.UserMobile, photoMallList)
		space.POST("/photo/top/set", authSvc.UserMobile, photoTopSet)
		space.GET("/photo/arc/list", verifySvc.Verify, authSvc.UserMobile, photoArcList)
		space.POST("/reserve", authSvc.UserMobile, reserve)
		space.POST("/reserve/cancel", authSvc.UserMobile, reserveCancel)
		space.POST("/reserve/upCancel", authSvc.UserMobile, upReserveCancel)
		space.POST("/reserve/shareInfo", authSvc.UserMobile, reserveShareInfo)
		spaceGarb := space.Group("/garb", featureSvc.BuildLimitHttp())
		{
			spaceGarb.GET("/detail", authSvc.GuestMobile, garbDetail)
			spaceGarb.GET("/list", authSvc.GuestMobile, userGarbList)
			spaceGarb.POST("/dress", proxyHandler, authSvc.UserMobile, garbDress)
			spaceGarb.POST("/take_off", proxyHandler, authSvc.UserMobile, garbTakeoff)
		}
		character := space.Group("/character", featureSvc.BuildLimitHttp())
		{
			character.GET("/list", authSvc.GuestMobile, characterList)
			character.POST("/set", proxyHandler, authSvc.UserMobile, characterSet)
			character.POST("/remove", proxyHandler, authSvc.UserMobile, characterRemove)
		}
		digital := space.Group("/digital", featureSvc.BuildLimitHttp())
		{
			digital.GET("/info", authSvc.GuestMobile, digitalInfo)
			digital.POST("/bind", proxyHandler, authSvc.UserMobile, digitalBind)
			digital.POST("/unbind", proxyHandler, authSvc.UserMobile, digitalUnbind)
			digital.GET("/extra/info", authSvc.GuestMobile, digitalExtraInfo)
		}
	}

	display := e.Group("/x/v2/display", verifySvc.Verify, featureSvc.BuildLimitHttp())
	display.GET("/zone", zone)
	display.GET("/id", authSvc.GuestMobile, displayID)

	favorite := e.Group("/x/v2/favorite", verifySvc.Verify, featureSvc.BuildLimitHttp())
	favorite.GET("", authSvc.GuestMobile, folder)
	favorite.GET("/video", authSvc.GuestMobile, favoriteVideo)
	favorite.GET("/topic", authSvc.GuestMobile, topic)
	favorite.GET("/article", authSvc.GuestMobile, article)
	favorite.GET("/clips", authSvc.GuestMobile, favClips)
	favorite.GET("/albums", authSvc.GuestMobile, favAlbums)
	favorite.GET("/sp", specil)
	favorite.GET("/audio", authSvc.GuestMobile, audio)
	favorite.GET("/tab", authSvc.UserMobile, tab)
	favorite.GET("/second/tab", authSvc.UserMobile, secondTab)
	favorite.GET("/channel", authSvc.GuestMobile, channel)

	relation := e.Group("/x/v2/relation", featureSvc.BuildLimitHttp())
	relation.GET("/followings", authSvc.GuestMobile, followings)
	relation.GET("/tag", authSvc.UserMobile, tag)
	relation.POST("/esport/add", authSvc.UserMobile, esportAdd)
	relation.POST("/esport/cancel", authSvc.UserMobile, esportCancel)

	history := e.Group("/x/v2/history", verifySvc.Verify, featureSvc.BuildLimitHttp())
	history.GET("", authSvc.UserMobile, historyList)
	history.GET("/live", live)
	history.GET("/liveList", authSvc.UserMobile, liveList)
	history.GET("/cursor", authSvc.UserMobile, historyCursor)
	history.POST("/del", proxyHandler, authSvc.UserMobile, historyDel)
	history.POST("/clear", proxyHandler, authSvc.UserMobile, historyClear)

	dataflow := e.Group("/x/v2/dataflow", featureSvc.BuildLimitHttp())
	dataflow.POST("/report", reportInfoc)

	account := e.Group("/x/v2/account", featureSvc.BuildLimitHttp())
	account.Use(bm.CORS())
	account.GET("/myinfo", verifySvc.Verify, myinfo)
	account.GET("/mine", verifySvc.Verify, authSvc.GuestMobile, mine)
	account.GET("/mine/ipad", verifySvc.Verify, authSvc.GuestMobile, mineIpad)
	account.POST("/config/set", verifySvc.Verify, authSvc.GuestMobile, configSet)
	account.GET("/export/statistics", authSvc.User, exportStatistics)
	account.GET("/nft/setting/button", verifySvc.Verify, authSvc.GuestMobile, nftSettingButton)

	account.POST("/teenagers/pwd/set", authSvc.UserMobile, teenagersPwd)
	account.GET("/teenagers/status", proxyHandler, authSvc.UserMobile, teenagersStatus)
	account.POST("/lessons/update", proxyHandler, authSvc.UserMobile, lessonsUpdate)
	account.POST("/teenagers/update", proxyHandler, authSvc.GuestMobile, teenagersUpdate)
	account.GET("/mode/status", proxyHandler, authSvc.GuestMobile, abtest.Handler(), modeStatus)
	account.POST("/teenagers/timer/set", proxyHandler, authSvc.GuestMobile, setTimer)
	account.GET("/teenagers/timer/get", proxyHandler, authSvc.GuestMobile, getTimer)

	account.POST("/bili_link/report", proxyHandler, authSvc.GuestMobile, biliLinkReport)

	family := e.Group("/x/v2/family", featureSvc.BuildLimitHttp(), proxyHandler)
	family.Use(bm.CORS())
	family.GET("/aggregation", authSvc.Guest, familyAggregation)
	family.GET("/teen_guard", authSvc.Guest, familyTeenGuard)
	family.GET("/identity", authSvc.User, familyIdentity)
	family.POST("/qrcode/create", authSvc.User, createFamilyQrcode)
	family.GET("/qrcode/info", authSvc.User, familyQrcodeInfo)
	family.GET("/qrcode/status", authSvc.User, familyQrcodeStatus)
	family.GET("/parent/index", authSvc.User, familyParentIndex)
	family.POST("/parent/unbind", authSvc.User, familyParentUnbind)
	family.POST("/parent/teenager/update", authSvc.User, parentUpdateTeenager)
	family.GET("/child/index", authSvc.User, familyChildIndex)
	family.POST("/child/bind", authSvc.User, familyChildBind)
	family.POST("/child/unbind", authSvc.User, familyChildUnbind)
	family.GET("/timelock/info", authSvc.User, timelockInfo)
	family.POST("/timelock/update", authSvc.User, updateTimelock)
	family.GET("/timelock/pwd", authSvc.User, timelockPwd)
	family.POST("/timelock/pwd/verify", authSvc.User, verifyTimelockPwd)

	addiction := e.Group("/x/v2/anti_addiction", featureSvc.BuildLimitHttp())
	addiction.GET("/rule", authSvc.User, antiAddictionRule)
	addiction.GET("/aggregation/status", proxyHandler, authSvc.GuestMobile, aggregationStatus)
	addiction.POST("/sleep_remind/set", proxyHandler, authSvc.User, setSleepRemind)
}

func internalRouter(e *bm.Engine) {
	e.Use(anticrawler.Report())
	group := e.Group("/x/internal/v2/space")
	{
		digital := group.Group("/digital")
		{
			digital.POST("/bind", digitalBind)
			digital.POST("/unbind", digitalUnbind)
		}
	}
}
