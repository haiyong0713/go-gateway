package http

import (
	"encoding/json"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/antispam"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/cache"
	"go-common/library/net/http/blademaster/middleware/cache/store"
	"go-common/library/net/http/blademaster/middleware/proxy"
	"go-common/library/net/http/blademaster/middleware/supervisor"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/metadata"
	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/space/interface/model"
	"time"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/middle"
	"go-gateway/app/web-svr/space/interface/service"
	"go-gateway/pkg/idsafe/bvid"
)

var (
	authSvr   *auth.Auth
	spcSvc    *service.Service
	spvSvc    *supervisor.Supervisor
	vfySvc    *verify.Verify
	antispamM *antispam.Antispam
	middleSvc *middle.Middle
	// databus
	visitPub *databus.Databus
	// cache components
	cacheSvr *cache.Cache
	deg      *cache.Degrader
)

// Init init http server
func Init(c *conf.Config, s *service.Service) {
	authSvr = auth.New(c.Auth)
	spvSvc = supervisor.New(c.Supervisor)
	vfySvc = verify.New(c.Verify)
	antispamM = antispam.New(c.Antispam)
	middleSvc = middle.New(c)
	visitPub = databus.New(c.Databus.VisitPub)
	spcSvc = s
	cacheSvr = cache.New(store.NewMemcache(c.DegradeConfig.Memcache))
	deg = cache.NewDegrader(c.DegradeConfig.Expire)
	// init http server
	engine := bm.DefaultServer(c.HTTPServer)
	outerRouter(engine)
	internalRouter(engine)
	// init Outer serve
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

func outerRouter(e *bm.Engine) {
	e.Use(bm.CORS(), anticrawler.Report())
	e.GET("/monitor/ping", ping)
	proxyHandler := proxy.NewZoneProxy("sh004", "http://sh001-api.bilibili.com")
	group := e.Group("/x/space")
	{
		abTest := group.Group("/abtest")
		{
			abTest.GET("", authSvr.Guest, abArcSearch)
		}
		chGroup := group.Group("/channel")
		{
			chGroup.GET("", authSvr.Guest, channel)
			chGroup.GET("/index", authSvr.Guest, channelIndex)
			chGroup.GET("/list", authSvr.Guest, channelList)
			chGroup.POST("/add", proxyHandler, spvSvc.ServeHTTP, authSvr.User, antispamM.ServeHTTP, middleSvc.Ban, addChannel)
			chGroup.POST("/del", proxyHandler, authSvr.User, antispamM.ServeHTTP, delChannel)
			chGroup.POST("/edit", proxyHandler, spvSvc.ServeHTTP, authSvr.User, antispamM.ServeHTTP, middleSvc.Ban, editChannel)
		}
		chvGroup := group.Group("/channel/video")
		{
			chvGroup.GET("", authSvr.Guest, channelVideo)
			chvGroup.POST("/add", proxyHandler, authSvr.User, antispamM.ServeHTTP, addChannelVideo)
			chvGroup.POST("/del", proxyHandler, authSvr.User, antispamM.ServeHTTP, delChannelVideo)
			chvGroup.POST("/sort", proxyHandler, authSvr.User, antispamM.ServeHTTP, sortChannelVideo)
			chvGroup.GET("/check", authSvr.User, checkChannelVideo)
		}
		riderGroup := group.Group("/rider")
		{
			riderGroup.GET("/list", authSvr.User, riderList)
			riderGroup.POST("/exit", authSvr.User, antispamM.ServeHTTP, exitRider)
		}
		tagGroup := group.Group("/tag")
		{
			tagGroup.POST("/sub", authSvr.User, antispamM.ServeHTTP, tagSub)
			tagGroup.POST("/sub/cancel", authSvr.User, antispamM.ServeHTTP, tagCancelSub)
			tagGroup.GET("/sub/list", authSvr.Guest, tagSubList)
		}
		bgmGroup := group.Group("/bangumi")
		{
			bgmGroup.POST("/concern", proxyHandler, authSvr.User, antispamM.ServeHTTP, bangumiConcern)
			bgmGroup.POST("/unconcern", proxyHandler, authSvr.User, antispamM.ServeHTTP, bangumiUnConcern)
			bgmGroup.GET("/concern/list", authSvr.Guest, bangumiList)
			bgmGroup.GET("/follow/list", authSvr.Guest, followList)
		}
		topGroup := group.Group("/top")
		{
			topGroup.GET("/arc", authSvr.Guest, topArc)
			topGroup.POST("/arc/set", proxyHandler, spvSvc.ServeHTTP, authSvr.User, antispamM.ServeHTTP, middleSvc.Ban, setTopArc)
			topGroup.POST("/arc/cancel", proxyHandler, authSvr.User, antispamM.ServeHTTP, cancelTopArc)
			topGroup.POST("/dynamic/set", proxyHandler, authSvr.User, antispamM.ServeHTTP, setTopDynamic)
			topGroup.POST("/dynamic/cancel", proxyHandler, authSvr.User, antispamM.ServeHTTP, cancelTopDynamic)
		}
		mpGroup := group.Group("/masterpiece")
		{
			mpGroup.GET("", authSvr.Guest, masterpiece)
			mpGroup.POST("/add", proxyHandler, spvSvc.ServeHTTP, authSvr.User, antispamM.ServeHTTP, middleSvc.Ban, addMasterpiece)
			mpGroup.POST("/edit", proxyHandler, spvSvc.ServeHTTP, authSvr.User, antispamM.ServeHTTP, middleSvc.Ban, editMasterpiece)
			mpGroup.POST("/cancel", proxyHandler, authSvr.User, antispamM.ServeHTTP, cancelMasterpiece)
		}
		noticeGroup := group.Group("/notice")
		{
			noticeGroup.GET("", notice)
			noticeGroup.POST("/set", proxyHandler, spvSvc.ServeHTTP, authSvr.User, antispamM.ServeHTTP, middleSvc.Ban, setNotice)
		}
		accGroup := group.Group("/acc")
		{
			accGroup.GET("/info", authSvr.Guest, accInfo)
			accGroup.GET("/tags", accTags)
			accGroup.POST("/tags/set", spvSvc.ServeHTTP, authSvr.User, antispamM.ServeHTTP, middleSvc.Ban, setAccTags)
			accGroup.GET("/relation", authSvr.User, relation)
		}
		themeGroup := group.Group("theme")
		{
			themeGroup.GET("/list", authSvr.User, themeList)
			themeGroup.POST("/active", proxyHandler, authSvr.User, antispamM.ServeHTTP, themeActive)
		}
		appGroup := group.Group("/app")
		{
			appGroup.GET("/index", authSvr.Guest, appIndex)
			appGroup.GET("/dynamic/list", authSvr.Guest, dynamicList)
			appGroup.GET("/played/game", authSvr.Guest, appPlayedGame)
			appGroup.GET("/top/photo", authSvr.Guest, appTopPhoto)
			appGroup.GET("/behavior/list", authSvr.Guest, behaviorList)
		}
		arcGroup := group.Group("/arc")
		{
			args := []string{"mid", "tid", "order", "keyword", "pn", "ps", "check_type", "check_id", "index"}
			arcGroup.GET("/search", authSvr.Guest, cacheSvr.Cache(deg.Args(args...), nil), arcSearch)
			arcGroup.GET("/list", arcList)
		}
		reserveGroup := group.Group("/reserve")
		{
			reserveGroup.POST("", authSvr.User, reserve)
			reserveGroup.POST("/cancel", authSvr.User, reserveCancel)
			reserveGroup.POST("/upCancel", authSvr.User, upReserveCancel)
		}
		dynamicGroup := group.Group("/dynamic")
		{
			dynamicGroup.GET("/search", authSvr.Guest, dynamicSearch)
		}
		topPhoto := group.Group("/topphoto")
		{
			topPhoto.GET("", authSvr.Guest, topPhotoIndex)
			topPhoto.GET("/mall", authSvr.User, topPhotoMallIndex)
			topPhoto.POST("/upload", authSvr.User, uploadTopPhoto)
			topPhoto.POST("/set", authSvr.User, setTopPhoto)
		}
		group.GET("/setting", settingInfo)
		group.GET("/article", article)
		group.GET("/navnum", authSvr.Guest, navNum)
		group.GET("/upstat", authSvr.Guest, upStat)
		group.GET("/shop", authSvr.User, shopInfo)
		group.GET("/album/index", albumIndex)
		group.GET("/fav/nav", authSvr.Guest, favNav)
		group.GET("/fav/arc", authSvr.Guest, favArc)
		group.GET("/fav/season/list", favSeasonList)
		group.GET("/coin/video", authSvr.Guest, coinVideo)
		group.GET("/like/video", authSvr.Guest, likeVideo)
		group.GET("/myinfo", authSvr.User, myInfo)
		group.GET("/lastplaygame", authSvr.Guest, lastPlayGame)
		group.POST("/privacy/modify", proxyHandler, authSvr.User, antispamM.ServeHTTP, privacyModify)
		group.POST("/index/order/modify", proxyHandler, authSvr.User, antispamM.ServeHTTP, indexOrderModify)
		group.POST("/privacy/batch/modify", proxyHandler, authSvr.User, antispamM.ServeHTTP, privacyBatchModify)
		group.GET("/privacy", authSvr.User, privacySetting)
		group.GET("/setting/app", authSvr.User, appSetting)
		group.GET("/activity/tab", authSvr.Guest, activityTab)
		group.GET("/reservation", authSvr.Guest, reservation)
	}
}

func internalRouter(e *bm.Engine) {
	e.Use(anticrawler.Report())
	group := e.Group("/x/internal/space")
	{
		group.GET("/setting", vfySvc.Verify, settingInfo)
		group.GET("/myinfo", vfySvc.Verify, authSvr.User, myInfo)
		group.POST("/privacy/modify", authSvr.User, privacyModify)
		group.POST("/privacy/batch/modify", authSvr.User, privacyBatchModify)
		group.POST("/index/order/modify", authSvr.User, indexOrderModify)
		accGroup := group.Group("/acc")
		{
			accGroup.GET("/info", vfySvc.Verify, authSvr.Guest, accInfo)
		}
		appGroup := group.Group("/app")
		{
			appGroup.GET("/index", vfySvc.Verify, authSvr.Guest, appIndex)
		}
		group.GET("/web/index", vfySvc.Verify, authSvr.Guest, webIndex)
		group.POST("/cache/clear", clearCache)
		group.GET("/blacklist", vfySvc.Verify, blacklist)
		group.GET("/system/notice", vfySvc.Verify, sysNotice)
		group.POST("/clear/msg", vfySvc.Verify, clearMsg)
		group.POST("/clear/topphoto/arc", vfySvc.Verify, clearTopPhotoArc)
		group.GET("/topphoto/arc", vfySvc.Verify, topPhotoArc)
		group.POST("/topphoto/arc/set", vfySvc.Verify, setTopPhotoArc)
		chvGroup := group.Group("/channel")
		{
			chvGroup.GET("/detail", channelDetail)
			chvGroup.GET("/video/aids", channelAids)
		}
		group.POST("/topphoto/cache/clear", vfySvc.Verify, clearCacheTopPhoto)
		group.POST("/member/cache/purge", purgeCacheTopPhoto)
	}
}

func ping(c *bm.Context) {}

func bvArgCheck(aid int64, bv string) (res int64, err error) {
	res = aid
	if bv != "" {
		if res, err = bvid.BvToAv(bv); err != nil {
			log.Error("bvid.BvToAv(%s) aid(%d) error(%+v)", bv, aid, err)
			err = ecode.RequestErr
			return
		}
	}
	if res <= 0 {
		err = ecode.RequestErr
	}
	return
}

func getRiskCommonReq(c *bm.Context, query string) (riskParams *model.RiskManagement) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	now := time.Now()
	header, _ := json.Marshal(c.Request.Header)
	riskParams = &model.RiskManagement{
		Mid:      mid,
		Buvid:    reqBuvid(c),
		Ip:       metadata.String(c, metadata.RemoteIP),
		Platform: "pc",
		Ctime:    now.Format("2006-01-02 15:04:05"),
		Api:      c.Request.URL.Path,
		Referer:  c.Request.Referer(),
		Ua:       c.Request.Header.Get("User-Agent"),
		Host:     c.Request.Host,
		Query:    query,
		Header:   string(header),
		Cookie:   c.Request.Header.Get("Cookie"),
	}
	return
}
