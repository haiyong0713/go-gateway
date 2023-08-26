package http

import (
	"encoding/json"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"
	"net/http"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/cache"
	"go-common/library/net/http/blademaster/middleware/cache/store"
	"go-common/library/net/http/blademaster/middleware/verify"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/service"
	channelSvr "go-gateway/app/web-svr/web/interface/service/channel"
	"go-gateway/pkg/idsafe/bvid"
)

var (
	webSvc  *service.Service
	authSvr *auth.Auth
	vfySvr  *verify.Verify
	chSvr   *channelSvr.Service

	// cache components
	cacheSvr       *cache.Cache
	deg            *cache.Degrader
	searchCacheSvr *cache.Cache
	searchDeg      *cache.Degrader
)

// Init init
func Init(c *conf.Config, s *service.Service) {
	authSvr = auth.New(c.Auth)
	vfySvr = verify.New(c.Verify)
	chSvr = channelSvr.New(c)
	webSvc = s
	cacheSvr = cache.New(store.NewMemcache(c.DegradeConfig.Memcache))
	deg = cache.NewDegrader(c.DegradeConfig.Expire)
	searchCacheSvr = cache.New(store.NewMemcache(c.SearchDegradeCache.Memcache))
	searchDeg = cache.NewDegrader(c.SearchDegradeCache.Expire)
	// init outer router
	engine := bm.NewServer(c.HTTPServer)
	engine.Use(bm.Recovery(), bm.Logger(), bm.Trace(), bm.Mobile(), bm.AuroraHandler())
	outerRouter(engine)
	internalRouter(engine)
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

func outerRouter(e *bm.Engine) {
	e.Use(bm.CORS(), anticrawler.Report())
	e.Ping(ping)
	e.GET("/x/web-interface/view", authSvr.Guest, view)
	group := e.Group("/x/web-interface", bm.CSRF())
	{
		arcGroup := group.Group("/archive")
		{
			arcGroup.GET("/coins", authSvr.User, coins)
			arcGroup.GET("/stat", archiveStat)
			arcGroup.GET("/desc", description)
			arcGroup.GET("/desc2", desc2)
			arcGroup.POST("/report", authSvr.User, arcReport)
			arcGroup.POST("/appeal", authSvr.User, arcAppeal)
			arcGroup.GET("/appeal/tags", appealTags)
			arcGroup.GET("/author/recommend", authorRecommend)
			arcGroup.GET("/related", authSvr.Guest, relatedArcs)
			arcGroup.POST("/like", authSvr.User, like)
			arcGroup.POST("/like/triple", authSvr.User, likeTriple)
			arcGroup.GET("/has/like", authSvr.User, hasLike)
			arcGroup.GET("/ugc/pay", authSvr.User, arcUGCPay)
			arcGroup.GET("/relation", authSvr.User, arcRelation)
			arcGroup.GET("/special/recommend", arcSpecRcmd)
			arcGroup.GET("/custom/config", arcCustomConfig)
		}
		dyGroup := group.Group("/dynamic")
		{
			dyGroup.GET("/region", dynamicRegion)
			dyGroup.GET("/index", dynamicRegions)
			dyGroup.GET("/tag", dynamicRegionTag)
			dyGroup.GET("/total", dynamicRegionTotal)
			dyGroup.GET("/entrance", authSvr.User, dynamicEntrance)
			dyGroup.GET("/card/type", authSvr.User, dynamicCardType)
			dyGroup.GET("/card/can/add/content", authSvr.User, dynamicCardCanAddContent)
			dyGroup.GET("/card/add", authSvr.User, dynamicCardAdd)
		}
		rankGroup := group.Group("/ranking")
		{
			rankGroup.GET("", ranking)
			rankGroup.GET("/v2", rankingV2)
			rankGroup.GET("/index", rankingIndex)
			rankGroup.GET("/region", rankingRegion)
			rankGroup.GET("/recommend", rankingRecommend)
			rankGroup.GET("/tag", rankingTag)
		}
		tagGroup := group.Group("/tag")
		{
			tagGroup.GET("/top", tagAids)
			tagGroup.GET("/archives", tagArchives)
		}
		artGroup := group.Group("/article")
		{
			artGroup.GET("/list", authSvr.Guest, articleList)
			artGroup.GET("/up/list", authSvr.Guest, articleUpList)
			artGroup.GET("/categories", categories)
			artGroup.GET("/newcount", newCount)
			artGroup.GET("/early", upMoreArts)
		}
		coinGroup := group.Group("/coin")
		{
			coinGroup.POST("/add", authSvr.User, addCoin)
			coinGroup.GET("/today/exp", authSvr.User, coinExp)
		}
		onlineGroup := group.Group("/online")
		{
			onlineGroup.GET("", onlineInfo)
			onlineGroup.GET("/list", onlineList)
			onlineGroup.GET("/total", onlineTotal)
		}
		helpGroup := group.Group("/help")
		{
			helpGroup.GET("/list", cacheSvr.Cache(deg.Args("parentTypeId"), nil), helpList)
			helpGroup.GET("/detail", cacheSvr.Cache(deg.Args("pn", "ps", "fId", "questionTypeId"), nil), helpDetail)
			helpGroup.GET("/search", helpSearch)
		}
		viewGroup := group.Group("/view")
		{
			viewGroup.GET("/detail", authSvr.Guest, detail)
			viewGroup.GET("/detail/tag", authSvr.Guest, detailTag)
			viewGroup.POST("/dm/vote", authSvr.User, dmVote)
			viewGroup.GET("/uplikeimg", authSvr.Guest, upLikeImg)
			viewGroup.GET("/premiere", authSvr.Guest, arcPremiere)
			viewGroup.GET("/premiere_info", authSvr.Guest, arcPremiereInfo)
		}
		searchGroup := group.Group("/search")
		{
			searchArgs := []string{"page", "keyword", "tids", "duration", "from_source", "highlight", "single_column", "dynamic_offset", "page_size"}
			searchGroup.GET("/all", authSvr.Guest, searchCacheSvr.Cache(searchDeg.Args(searchArgs...), nil), searchAll)
			searchGroup.GET("/all/v2", authSvr.Guest, searchCacheSvr.Cache(searchDeg.Args(searchArgs...), nil), searchAllV2)
			searchGroup.GET("/type", authSvr.Guest, searchCacheSvr.Cache(searchDeg.Args("page", "search_type", "keyword", "order", "tids", "from_source", "platform", "duration", "category_id", "vp_num", "bili_user_vl", "user_type", "order_sort", "highlight", "single_column", "dynamic_offset", "page_size"), nil), searchByType)
			searchGroup.GET("/recommend", authSvr.Guest, searchRec)
			searchGroup.GET("/default", authSvr.Guest, searchDefault)
			searchGroup.GET("/egg", searchEgg)
			searchGroup.GET("/game/info", searchGameInfo)
			searchGroup.GET("/square", authSvr.Guest, searchSquare)
		}
		wxGroup := group.Group("/wx")
		{
			wxGroup.GET("/hot", wxHot)
			wxGroup.GET("/search/all", authSvr.Guest, wxSearchAll)
			wxGroup.GET("/history/cursor", authSvr.User, wxHistoryCursor)
		}
		bnjGroup := group.Group("/bnj2019")
		{
			bnjGroup.GET("", authSvr.Guest, bnj2019)
			bnjGroup.GET("/timeline", authSvr.Guest, timeline)
		}
		bnj20Group := group.Group("/bnj2020")
		{
			bnj20Group.GET("", authSvr.Guest, bnj2020)
			bnj20Group.GET("/item", authSvr.Guest, bnj2020Item)
			bnj20Group.GET("/elec/show", bnj2020ElecShow)
			bnj20Group.GET("/timeline", authSvr.Guest, bnj2020Timeline)
			bnj20Group.GET("/aids", bnj2020Aids)
		}
		indexGroup := group.Group("/index/sort")
		{
			indexGroup.GET("", authSvr.User, indexSort)
			indexGroup.POST("/set", authSvr.User, indexSet)
		}
		vlogGroup := group.Group("/vlog", authSvr.Guest)
		{
			vlogGroup.GET("", vlog)
			vlogGroup.GET("/rank", vlogRank)
		}
		channelGroup := group.Group("/channel", authSvr.Guest)
		{
			channelGroup.GET("/detail", channelDetail)
			channelGroup.GET("/multiple", channelMultiple)
			channelGroup.GET("/selected", channelSelected)
		}
		gwGroup := group.Group("/gateway")
		{
			gwGroup.GET("/dynamic/material/info", materialInfo)
		}
		chGroup := group.Group("/web/channel")
		{
			chGroup.GET("/red", authSvr.Guest, channelRed)
			chGroup.GET("/category/list", categoryList)
			chGroup.GET("/category/channel/list", authSvr.Guest, channelList)
			chGroup.GET("/category/channel_arc/list", authSvr.Guest, channelArcList)
			chGroup.GET("/subscribe/list", authSvr.User, subscribedList)
			chGroup.GET("/view/list", authSvr.User, viewList)
			chGroup.POST("/stick", authSvr.User, stick)
			chGroup.POST("/subscribe", authSvr.User, subscribe)
			chGroup.POST("/unsubscribe", authSvr.User, unsubscribe)
			chGroup.GET("/hot/list", authSvr.Guest, hotList)
			chGroup.GET("/detail", authSvr.Guest, webDetail)
			chGroup.GET("/featured/list", authSvr.Guest, featuredList)
			chGroup.GET("/multiple/list", authSvr.Guest, multipleList)
			chGroup.GET("/search", authSvr.Guest, searchChannel)
			chGroup.GET("/top/list", authSvr.Guest, topList)
		}
		popularGroup := group.Group("/popular")
		{
			popularGroup.GET("", authSvr.Guest, webPopular)
			popularGroup.GET("/series/list", popularSeries)
			popularGroup.GET("/series/one", popularSeriesOne)
			popularGroup.GET("/precious", popularPrecious)
			// 入站必刷活动
			popularGroup.GET("/precious/activity", authSvr.User, popularActivity)
			popularGroup.GET("/precious/activity/archive/list", authSvr.User, popularActivityArchiveList)
			popularGroup.POST("/precious/activity/award", authSvr.User, popularActivityAward)
		}
		activityGroup := group.Group("/activity/season")
		{
			activityGroup.GET("", authSvr.Guest, activitySeason)
			activityGroup.GET("/archive", authSvr.Guest, activityArchive)
			activityGroup.GET("/live/time", authSvr.Guest, activityLiveTimeInfo)
			activityGroup.POST("/click", authSvr.User, activitySeasonClick)
		}
		lpGroup := group.Group("/landing/page")
		{
			lpGroup.GET("/newlist", authSvr.Guest, lpNewList)
			lpGroup.GET("/dynamic/region", lpDynamicRegion)
			lpGroup.GET("/ranking/recommend", lpRankingRecommend)
		}
		group.GET("/attentions", authSvr.User, attentions)
		group.GET("/card", authSvr.Guest, card)
		group.GET("/nav", authSvr.Guest, nav)
		group.GET("/nav/stat", authSvr.User, navStat)
		group.GET("/newlist", newList)
		group.GET("/information", authSvr.Guest, information) // 资讯区接口 效果同newlist
		group.POST("/feedback", authSvr.Guest, feedback)
		group.GET("/zone", ipZone)
		group.POST("/share/add", authSvr.Guest, addShare)
		group.GET("/elec/show", elecShow)
		group.GET("/index/icon", indexIcon)
		group.GET("/baidu/kv", kv)
		group.GET("/cmtbox", cmtbox)
		group.GET("/abserver", authSvr.Guest, abServer)
		group.GET("/up/rec", authSvr.User, upRec)
		group.GET("/broadcast/servers", broadServer)
		group.GET("/index/top", webTop)
		group.GET("/index/top/rcmd", authSvr.Guest, webTopRcmd)
		group.GET("/index/top/feed/rcmd", authSvr.Guest, webTopFeedRcmdV2)
		group.GET("/history/cursor", authSvr.User, historyCursor)
		group.GET("/relation", authSvr.User, relation)
		group.GET("/cdn/report", cdnReport)
		group.GET("/param", paramConfig)
		group.GET("/region/index", regionIndex)
		group.GET("/activity/movie/review/list", activityMovieList) // 电影评分活动
		group.GET("/h5onelink", getOnelink)                         // 海外h5区分来源
		appealGroup := group.Group("/pwd_appeal")
		{
			appealGroup.POST("/add", authSvr.Guest, addPwdAppeal)
			appealGroup.POST("/upload", authSvr.Guest, uploadPwdAppeal)
			appealGroup.POST("/captcha/send", pwdAppealSendCaptcha)
		}
		campusGroup := group.Group("/campus")
		{
			campusGroup.GET("/index", authSvr.User, pages)
			campusGroup.GET("/school/search", schoolSearch)
			campusGroup.GET("/school/recommend", schoolRecommend)
			campusGroup.GET("/official/accounts", authSvr.User, OfficialAccounts)
			campusGroup.GET("/official/dynamics", authSvr.User, OfficialDynamics)
			campusGroup.GET("/topic/list", authSvr.User, CampusTopicList)
			campusGroup.GET("/billboard", authSvr.User, CampusBillboard)
			campusGroup.POST("/feedback", authSvr.User, CampusFeedback)
			campusGroup.GET("/nearby", authSvr.Guest, CampusNearbyRcmd)
			campusGroup.GET("/reddot", authSvr.User, CampusRedDot)
		}
		group.GET("/vas/trade/create", authSvr.User, tradeCreate) // 付费合集订单创建
		pcdnGroup := group.Group("/pcdn")
		{
			pcdnGroup.POST("/join", authSvr.User, joinPCDN)
			pcdnGroup.POST("/operate", authSvr.User, operatePCDN)
			pcdnGroup.GET("/user/settings", authSvr.User, userSettings)
			pcdnGroup.GET("/user/info", authSvr.User, userAccountInfo)
			pcdnGroup.POST("/exchange", authSvr.User, exchange)
			pcdnGroup.GET("/v1", authSvr.User, pcdnV1)
			pcdnGroup.POST("/report", authSvr.User, pcdnReport)
			pcdnGroup.GET("/notify", authSvr.User, notify)
			pcdnGroup.GET("/digital/collection", authSvr.User, digitialCollection)
			pcdnGroup.POST("/quit", authSvr.User, quit)
			pcdnGroup.GET("/pages", authSvr.User, pacnPages)
			// pcdnGroup.GET("/digital/exchange", authSvr.User, digitialCollection)
		}
		group.GET("bgroup/member/in", authSvr.User, memberIn) // 人群包
	}
	e.GET("/x/coin/list", coinList)
	e.GET("/serverdate.js", serverDate)
	e.GET("/plus/widget/ajaxGetCaptchaKey.php", captchaKey)
}

func internalRouter(e *bm.Engine) {
	e.Use(anticrawler.Report())
	group := e.Group("/x/internal/web-interface")
	{
		dyGroup := group.Group("/dynamic")
		{
			dyGroup.GET("/region", vfySvr.Verify, dynamicRegion)
			dyGroup.GET("/index", vfySvr.Verify, dynamicRegions)
			dyGroup.GET("/tag", vfySvr.Verify, dynamicRegionTag)
			dyGroup.GET("/total", vfySvr.Verify, dynamicRegionTotal)
		}
		rankGroup := group.Group("/ranking")
		{
			rankGroup.GET("", vfySvr.Verify, ranking)
			rankGroup.GET("/v2", vfySvr.Verify, rankingV2)
			rankGroup.GET("/index", vfySvr.Verify, rankingIndex)
			rankGroup.GET("/region", vfySvr.Verify, rankingRegion)
			rankGroup.GET("/recommend", vfySvr.Verify, rankingRecommend)
			rankGroup.GET("/tag", vfySvr.Verify, rankingTag)
		}
		tagGroup := group.Group("/tag")
		{
			tagGroup.GET("/top", vfySvr.Verify, tagAids)
			tagGroup.GET("/detail", vfySvr.Verify, tagDetail)
		}
		helpGroup := group.Group("/help")
		{
			helpGroup.GET("/list", vfySvr.Verify, helpList)
			helpGroup.GET("/detail", vfySvr.Verify, helpDetail)
			helpGroup.GET("/search", vfySvr.Verify, helpSearch)
		}
		onlineGroup := group.Group("/online")
		{
			onlineGroup.GET("", vfySvr.Verify, onlineInfo)
			onlineGroup.GET("/list", vfySvr.Verify, onlineList)
		}
		viewGroup := group.Group("/view")
		{
			viewGroup.GET("", vfySvr.Verify, authSvr.Guest, view)
			viewGroup.GET("/detail", vfySvr.Verify, authSvr.Guest, detail)
			viewGroup.GET("/uplikeimg", vfySvr.Verify, authSvr.Guest, upLikeImg)
			viewGroup.GET("/premiere", vfySvr.Verify, authSvr.Guest, arcPremiere)
			viewGroup.GET("/premiere_info", vfySvr.Verify, authSvr.Guest, arcPremiereInfo)
		}
		searchGroup := group.Group("/search")
		{
			searchGroup.GET("/all", vfySvr.Verify, authSvr.Guest, searchAll)
			searchGroup.GET("/all/v2", vfySvr.Verify, authSvr.Guest, searchAllV2)
			searchGroup.GET("/type", vfySvr.Verify, authSvr.Guest, searchByType)
			searchGroup.GET("/recommend", vfySvr.Verify, authSvr.Guest, searchRec)
			searchGroup.GET("/square", vfySvr.Verify, authSvr.Guest, searchSquare)
		}
		bnj20Group := group.Group("/bnj2020")
		{
			bnj20Group.GET("", vfySvr.Verify, authSvr.Guest, bnj2020)
			bnj20Group.GET("/aids", vfySvr.Verify, bnj2020Aids)
		}
		popularGroup := group.Group("/popular")
		{
			popularGroup.GET("", vfySvr.Verify, authSvr.Guest, webPopular)
		}
		activityGroup := group.Group("/activity/season")
		{
			activityGroup.GET("", vfySvr.Verify, authSvr.Guest, activitySeason)
			activityGroup.GET("/archive", vfySvr.Verify, authSvr.Guest, activityArchive)
			activityGroup.GET("/live/time", vfySvr.Verify, authSvr.Guest, activityLiveTimeInfo)
		}
		channelGroup := group.Group("/channel", vfySvr.Verify)
		{
			channelGroup.GET("/detail", authSvr.Guest, channelDetail)
		}
		chGroup := group.Group("/web/channel", vfySvr.Verify)
		{
			chGroup.GET("/category/channel/list", authSvr.Guest, channelList)
			chGroup.GET("/category/channel_arc/list", authSvr.Guest, channelArcList)
			chGroup.GET("/view/list", authSvr.User, viewList)
			chGroup.GET("/detail", authSvr.Guest, webDetail)
			chGroup.GET("/multiple/list", authSvr.Guest, multipleList)
			chGroup.GET("/featured/list", authSvr.Guest, featuredList)
		}
		group.GET("/index/top", vfySvr.Verify, webTop)
		group.GET("/index/top/rcmd", vfySvr.Verify, authSvr.Guest, webTopRcmd)
		group.GET("/index/top/feed/rcmd", vfySvr.Verify, authSvr.Guest, webTopFeedRcmdV2)
		group.GET("/newlist", vfySvr.Verify, newList)
		group.GET("/zone", vfySvr.Verify, ipZone)
		group.GET("/baidu/kv", vfySvr.Verify, kv)
		group.GET("/cmtbox", vfySvr.Verify, cmtbox)
		group.GET("/broadcast/servers", vfySvr.Verify, broadServer)
		group.GET("/bnj2019", vfySvr.Verify, authSvr.Guest, bnj2019)
		group.GET("/bnj2019/aids", vfySvr.Verify, bnj2019Aids)
		group.GET("/index/sort", vfySvr.Verify, indexSort)
		group.GET("/av/config", vfySvr.Verify, avConfig)
		group.GET("/wx/hot", vfySvr.Verify, wxHot)
		group.GET("/region/index", vfySvr.Verify, regionIndex)
		group.GET("/h5onelink", vfySvr.Verify, getOnelink) // 海外h5区分来源
		lpGroup := group.Group("/landing/page")
		{
			lpGroup.GET("/newlist", authSvr.Guest, lpNewList)
			lpGroup.GET("/dynamic/region", lpDynamicRegion)
			lpGroup.GET("/ranking/recommend", lpRankingRecommend)
		}
		campusGroup := group.Group("/campus")
		{
			campusGroup.GET("/index", vfySvr.Verify, authSvr.User, pages)
			campusGroup.GET("/school/search", vfySvr.Verify, schoolSearch)
			campusGroup.GET("/school/recommend", vfySvr.Verify, schoolRecommend)
			campusGroup.GET("/official/accounts", vfySvr.Verify, authSvr.User, OfficialAccounts)
			campusGroup.GET("/official/dynamics", vfySvr.Verify, authSvr.User, OfficialDynamics)
			campusGroup.GET("/topic/list", vfySvr.Verify, authSvr.User, CampusTopicList)
			campusGroup.GET("/billboard", vfySvr.Verify, authSvr.User, CampusBillboard)
			campusGroup.POST("/feedback", vfySvr.Verify, authSvr.User, CampusFeedback)
			campusGroup.GET("/nearby", vfySvr.Verify, authSvr.Guest, CampusNearbyRcmd)
			campusGroup.GET("/reddot", vfySvr.Verify, authSvr.User, CampusRedDot)
		}
	}
}

func ping(c *bm.Context) {
	if err := webSvc.Ping(c); err != nil {
		log.Error("web-interface  ping error")
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func bvArgCheck(aid int64, bv string) (res int64, err error) {
	res = aid
	if bv != "" {
		if res, err = bvid.BvToAv(bv); err != nil {
			log.Error("bvArgCheck bvid.BvToAv(%s) aid(%d) error(%+v)", bv, aid, err)
			err = ecode.RequestErr
			return
		}
	}
	if res <= 0 {
		err = ecode.RequestErr
	}
	return
}

func cdnReport(c *bm.Context) {
	c.JSON(nil, nil)
}

func reqBuvid(ctx *bm.Context) string {
	buvid := ctx.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := ctx.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	return buvid
}

func reqSid(ctx *bm.Context) string {
	cookieSid, err := ctx.Request.Cookie("sid")
	if err != nil {
		return ""
	}
	return cookieSid.Value
}

func getRiskCommonReq(c *bm.Context, query string) (riskParams *model.RiskManagement) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	now := time.Now()
	header, _ := json.Marshal(c.Request.Header)
	riskParams = &model.RiskManagement{
		Mid:       mid,
		Buvid:     reqBuvid(c),
		Ip:        metadata.String(c, metadata.RemoteIP),
		Platform:  "pc",
		Ctime:     now.Format("2006-01-02 15:04:05"),
		Api:       c.Request.URL.Path,
		Referer:   c.Request.Referer(),
		UserAgent: c.Request.Header.Get("User-Agent"),
		Host:      c.Request.Host,
		Query:     query,
		Header:    string(header),
		Cookie:    c.Request.Header.Get("Cookie"),
		ItemType:  "av",
	}
	return
}
