package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"

	"go-gateway/app/app-svr/app-feed/admin/service/entry"
	"go-gateway/app/app-svr/app-feed/admin/service/push"
	"go-gateway/app/app-svr/app-feed/admin/service/pwd_appeal"
	"go-gateway/app/app-svr/app-feed/admin/service/search_whitelist"
	"go-gateway/app/app-svr/app-feed/admin/service/spmode"
	"go-gateway/app/app-svr/app-feed/admin/service/teen_manual"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/middle"
	"go-gateway/app/app-svr/app-feed/admin/service/aggregation"
	aggregation2 "go-gateway/app/app-svr/app-feed/admin/service/aggregation_v2"
	antisvr "go-gateway/app/app-svr/app-feed/admin/service/anti_crawler"
	bfssvr "go-gateway/app/app-svr/app-feed/admin/service/bfs"
	"go-gateway/app/app-svr/app-feed/admin/service/bubble"
	"go-gateway/app/app-svr/app-feed/admin/service/card"
	"go-gateway/app/app-svr/app-feed/admin/service/channel"
	"go-gateway/app/app-svr/app-feed/admin/service/common"
	"go-gateway/app/app-svr/app-feed/admin/service/egg"
	ftsvr "go-gateway/app/app-svr/app-feed/admin/service/feature"
	"go-gateway/app/app-svr/app-feed/admin/service/frontpage"
	"go-gateway/app/app-svr/app-feed/admin/service/hidden"
	"go-gateway/app/app-svr/app-feed/admin/service/icon"
	"go-gateway/app/app-svr/app-feed/admin/service/information"
	"go-gateway/app/app-svr/app-feed/admin/service/intervention"
	"go-gateway/app/app-svr/app-feed/admin/service/menu"
	pgcsvr "go-gateway/app/app-svr/app-feed/admin/service/pgc"
	"go-gateway/app/app-svr/app-feed/admin/service/popular"
	"go-gateway/app/app-svr/app-feed/admin/service/rank"
	"go-gateway/app/app-svr/app-feed/admin/service/resource"
	"go-gateway/app/app-svr/app-feed/admin/service/search"
	"go-gateway/app/app-svr/app-feed/admin/service/selected"
	"go-gateway/app/app-svr/app-feed/admin/service/sidebar"
	"go-gateway/app/app-svr/app-feed/admin/service/splash_screen"
	"go-gateway/app/app-svr/app-feed/admin/service/tianma"
	"go-gateway/app/app-svr/app-feed/admin/service/tips"
	"go-gateway/app/app-svr/app-feed/admin/service/web"
)

var (
	authSvc         *permit.Permit
	eggSvc          *egg.Service
	bfsSvc          *bfssvr.Service
	searchSvc       *search.Service
	searchWhitelist *search_whitelist.Service
	selSvc          *selected.Service
	pgcSvr          *pgcsvr.Service
	chanelSvc       *channel.Service
	popularSvc      *popular.Service
	infoSvc         *information.Service
	//nolint:unused
	cardSvc         *channel.Service
	commonSvc       *common.Service
	webSvc          *web.Service
	aggSvc          *aggregation.Service
	resourceSvc     *resource.Service
	menuSvr         *menu.Service
	vfySvr          *verify.Verify
	aggSvc2         *aggregation2.Service
	bubbleSvc       *bubble.Service
	middleSvc       *middle.Permit
	hiddenSvc       *hidden.Service
	iconSvc         *icon.Service
	sidebarSvc      *sidebar.Service
	interventionSrv *intervention.Service
	tianmaSvc       *tianma.Service
	splashSvc       *splash_screen.Service
	entrySvc        *entry.Service
	rankSvc         *rank.Service
	featureSvc      *ftsvr.Service
	tipsSvc         *tips.Service
	resourceCardSvc *card.Service
	frontpageSvc    *frontpage.Service
	pwdAppealSvc    *pwd_appeal.Service
	spmodeSvc       *spmode.Service
	pushSvc         *push.Service
	antiSvr         *antisvr.Service
	teenManualSvc   *teen_manual.Service
)

// initService init service
func initService(c *conf.Config, searchService *search.Service) {
	vfySvr = verify.New(nil)
	authSvc = permit.New2(nil)
	eggSvc = egg.New(c)
	bfsSvc = bfssvr.New(c)
	searchSvc = searchService
	searchWhitelist = search_whitelist.New(c)
	selSvc = selected.New(c)
	pgcSvr = pgcsvr.New(c)
	chanelSvc = channel.New(c)
	cardSvc = channel.New(c)
	popularSvc = popular.New(c)
	infoSvc = information.New(c)
	commonSvc = common.New(c)
	webSvc = web.New(c)
	aggSvc = aggregation.New(c)
	aggSvc2 = aggregation2.New(c)
	resourceSvc = resource.New(c)
	menuSvr = menu.New(c)
	bubbleSvc = bubble.New(c)
	middleSvc = middle.New(nil)
	hiddenSvc = hidden.New(c)
	iconSvc = icon.New(c)
	sidebarSvc = sidebar.New(c)
	interventionSrv = intervention.New(c)
	tianmaSvc = tianma.New(c)
	splashSvc = splash_screen.New(c)
	entrySvc = entry.New(c)
	rankSvc = rank.New(c)
	featureSvc = ftsvr.New(c)
	tipsSvc = tips.New(c)
	resourceCardSvc = card.New(c)
	frontpageSvc = frontpage.New(c)
	pwdAppealSvc = pwd_appeal.NewService(c)
	spmodeSvc = spmode.NewService(c)
	pushSvc = push.New(c)
	antiSvr = antisvr.New(c)
	teenManualSvc = teen_manual.NewService(c)
}

// Init init http sever instance.
func Init(c *conf.Config, searchService *search.Service) {
	initService(c, searchService)
	engine := bm.DefaultServer(c.HTTPServer)
	innerRouter(engine)
	// init internal server
	if err := engine.Start(); err != nil {
		log.Error("httpx.Serve error(%v)", err)
		panic(err)
	}
}

// innerRouter
func innerRouter(e *bm.Engine) {
	// ping monitor
	e.GET("/monitor/ping", ping)
	// modules color eggs
	feed := e.Group("/x/admin/feed")
	{
		feed.POST("/upload", clientUpload)
		feed.POST("/special/upload", specialUpload)
		// 对外 搜索
		feed.GET("/eggSearch", searchEgg)
		// 对外 web
		feed.GET("/eggSearchWeb", SearchEggWeb)
		common := feed.Group("/common")
		{
			common.GET("/card/titlePreview", cardPreview)
			common.GET("/card/titlePreview/batch", cardPreviewBatch)
			common.GET("/log/action", actionLog)
			common.POST("/log/addAction", actionAddLog)
			common.GET("/pgc/season", getPgcSeason)
			common.GET("/pgc/seasons", getPgcSeasons)
			common.GET("/pgc/ep", getPgcEp)
			common.GET("/card/type", cardType)
			common.GET("/archive/type", archiveType)
			common.GET("/tag/type", tagType)
			common.GET("/archives", archives)
			common.POST("/notify", notify)
			common.GET("/upinfo", upInfo)
			common.GET("/comicInfo", comicInfo)
			common.GET("/bvav", bvToAv)
			common.GET("/avbv", avToBv)
			common.GET("/gameInfo", gameInfo)
			common.GET("/game/app", gameInfoApp)
			common.POST("/upload/base64", uploadBase64)
			common.POST("/validate/gif", validatGif)
		}
		// 对外
		open := feed.Group("/open")
		{
			open.GET("/search/egg", searchEgg)                            // 搜索彩蛋
			open.GET("/search/webegg", SearchEggWeb)                      // 搜索web彩蛋
			open.POST("/search/addHotword", openAddHotword)               // 搜索 添加热词
			open.POST("/search/addDarkword", openAddDarkword)             // 搜索 添加黑马词
			open.GET("/search/blackList", openBlacklist)                  // 搜索 黑名单
			open.GET("/search/hotwords", openHotList)                     // 搜索 热词
			open.GET("/search/darkword", openDarkword)                    // 搜索 获取黑马词
			open.GET("/search/webSearch", openSearchWeb)                  // web 搜索
			open.GET("/search/shield", openSearchShield)                  // 搜索屏蔽
			open.GET("/search/web/module", openSearchModule)              // web 搜索模块
			open.POST("/ai/addPopStars", aiAddPopularStars)               // AI 添加新星卡片
			open.GET("/search/recommend", openRecommend)                  // 搜索web &&app 相关推荐
			open.GET("/search/ogv", openOgv)                              // 搜索web &&app 相关推荐
			open.GET("/splash/list", openSplashList)                      // 给网关用，上线中的闪屏配置
			open.GET("/search/channel/ids", openChannelIds)               // 给频道服务端用，返回管理后台所有配置过的频道id
			open.GET("/rank/list", openRankList)                          // 给网关用，返回所有用户可见的榜单信息 http://bapi.bilibili.co/project/3077/interface/api/cat_25197
			open.GET("/search/tips", openSearchTips)                      // 给网关和AI用，返回条件内搜索提示的配置 https://www.tapd.bilibili.co/20092951/prong/stories/view/1120092951001385200
			open.GET("/search/whiteList", openSearchWhiteList)            // 给网关和AI用，返回搜索结果的白名单视频 https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661001434364
			open.GET("/search/brand/blacklist", openSearchBrandBlacklist) // 给AI用，返回搜索品专黑名单列表
		}
		search := feed.Group("/search")
		{
			searchHot := search.Group("", authSvc.Permit2("SEARCH_HOTWORD"))
			{
				searchHot.GET("/blackList", blackList)
				searchHot.POST("/addBlack", addBlack)
				searchHot.POST("/delBlack", delBlack)
				searchHot.GET("/hot", hotList)
				searchHot.POST("/hot/sort", hotSort)                     // 对热词干预排序
				searchHot.GET("/hot/top", hotTop)                        // 获取在线的搜索热词的前20名
				searchHot.GET("/hot/pending", hotPending)                // 获取搜索热词的预定池
				searchHot.GET("/hot/statistics", hotStatistics)          // 获取搜索热词的历史数据
				searchHot.GET("/hot/statistics_live", hotStatisticsLive) // 获取搜索热词的历史数据
				searchHot.POST("/addInter", addInter)
				searchHot.GET("/intervene/history", interHistory)
				searchHot.POST("/updateInter", updateInter)
				searchHot.POST("/deleteHot", deleteHot)
				searchHot.POST("/updateSearch", updateSearch)
				searchHot.POST("/publishHot", publishHotWord)
				searchHot.POST("/publishDark", publishDarkWord)
				searchHot.GET("/dark", darkList)
				searchHot.POST("/delDark", deleteDark)
			}
			searchWeb := search.Group("/web", authSvc.Permit2(""))
			{
				searchWeb.GET("/card/list", searchWebCardList)
				searchWeb.POST("/card/add", addSearchWebCard)
				searchWeb.POST("/card/update", upSearchWebCard)
				searchWeb.POST("/card/delete", delSearchWebCard)
				searchWeb.GET("/list", searchWebList) // 搜索管理 -> web端配置
				searchWeb.POST("/add", addSearchWeb)
				searchWeb.POST("/update", upSearchWeb)
				searchWeb.POST("/delete", delSearchWeb)
				searchWeb.POST("/opt", optSearchWeb)
				searchWeb.POST("/batchOpt", batchOptWebRcmd) // 不知道是否正在使用，不能轻易改动，推荐使用下面/opt/batch接口
				searchWeb.POST("/opt/batch", batchOptSearchWeb)
				searchWeb.POST("/release", releaseSearchWeb) // 2021M9W1：为视频模块小卡发版时对老数据处理，兼容旧版本
			}
			searchEgg := search.Group("/egg")
			{
				searchEgg.POST("/add", addEgg)
				searchEgg.GET("/index", indexEgg)
				searchEgg.POST("/update", updateEgg)
				searchEgg.POST("/publish", pubEgg)
				searchEgg.POST("/delete", delEgg)
			}
			shield := search.Group("/shield", authSvc.Permit2("SEARCH_SHIELD"))
			{
				shield.GET("", searchShield)
				shield.POST("/add", addSearchShield)
				shield.POST("/update", upSearchShield)
				shield.POST("/option", optSearchShield)
			}
			spread := search.Group("/result/spread")
			{
				opt := spread.Group("/opt", authSvc.Permit2("SEARCH_CHECK_PRIV"))
				{
					opt.POST("/batch", batchOptSearchResultSpread)
				}
			}
			searchOgv := search.Group("/ogv", middleSvc.Permit("SEARCH_OGV,SEARCH_OGV_EDIT"))
			{
				searchOgv.GET("", ogvList)
				searchOgv.POST("/add", addOgv)
				searchOgv.POST("/update", updateOgv)
				searchOgv.POST("/opt", optOgv)
			}
			webModule := search.Group("/web/module", authSvc.Permit2("SEARCH_WEB_MODULE"))
			{
				webModule.GET("", searchWebModule)
				webModule.POST("/add", addWebModule)
				webModule.POST("/update", upSearchWebModule)
				webModule.POST("/option", optSearchWebModule)
			}
			// 搜索提示
			tips := search.Group("/tips", authSvc.Permit2("SEARCH_TIPS"))
			{
				tips.GET("", searchTipList)
				tips.POST("/add", searchTipAdd)
				tips.POST("/update", searchTipUpdate)
				tips.POST("/operate", searchTipOperate)
				tips.POST("/offline", searchTipOffline)
			}
			// 搜索白名单
			whiteList := search.Group("/whiteList")
			{
				whiteList.GET("/list", searchWhiteList)
				whiteList.GET("/archiveList", searchWhiteListArchiveList)
				whiteList.POST("/add", searchWhiteListAdd)
				whiteList.POST("/edit", searchWhiteListEdit)
				whiteList.POST("/option", searchWhiteListOption)
				whiteList.GET("/archivePreview", searchWhiteListArchivePreview)
			}
			// 搜索品专黑名单
			brand := search.Group("/brand", authSvc.Permit2("SEARCH_BRAND_BLACKLIST"))
			{
				brand.GET("/blacklist/list", searchBrandBlacklistList)
				brand.POST("/blacklist/add", searchBrandBlacklistAdd)
				brand.POST("/blacklist/edit", searchBrandBlacklistEdit)
				brand.POST("/blacklist/option", searchBrandBlacklistOption)
			}
			upAlias := search.Group("/up_alias")
			{
				upAlias.POST("/add", searchAddUpAlias, authSvc.Permit2("SEARCH_UP_ALIAS"))
				upAlias.POST("/edit", searchEditUpAlias, authSvc.Permit2("SEARCH_UP_ALIAS"))
				upAlias.POST("/toggle", searchToggleUpAlias, authSvc.Permit2("SEARCH_UP_ALIAS"))
				upAlias.GET("/list", searchSearchUpAlias, authSvc.Permit2("SEARCH_UP_ALIAS"))
				upAlias.GET("/export", searchExportUpAlias, authSvc.Permit2("SEARCH_UP_ALIAS"))
				upAlias.GET("/sync", syncSearchUpAlias)
			}
		}

		// http://bapi.bilibili.co/project/3077/interface/api/cat_25197
		rank := feed.Group("/rank")
		{

			rank.GET("/list", rankList)
			rank.GET("/detail", rankDetail)
			rank.GET("/config", rankConfig)
			rank.POST("/add", rankAdd)
			rank.POST("/edit", rankEdit)
			rank.POST("/publish", rankPublish)
			rank.POST("/terminate", rankTerminate)
			rank.POST("/change_state", rankOption)
			rank.POST("/manually_run_job", manuallyRunJob)

			archive := rank.Group("/archive")
			{
				archive.POST("/edit", rankArchiveEdit)
				archive.POST("/add", rankArchiveAdd)
			}

		}

		cardsetup := feed.Group("/cardsetup")
		{
			cardsetup.POST("/add", addCardSetup)
			cardsetup.GET("/list", cardSetupList)
			cardsetup.POST("/delete", delCardSetup)
			cardsetup.POST("/update", updateCardSetup)
		}
		channel := feed.Group("/channel")
		{
			tab := channel.Group("/tab")
			{
				tab.GET("/list", tabList)
				tab.POST("/add", addTab)
				tab.POST("/update", updateTab)
				tab.POST("/delete", deleteTab)
				tab.POST("/offline", offlineTab)
			}
		}
		popular := feed.Group("/popular")
		{
			eventTopic := popular.Group("/event_topic")
			{
				eventTopic.GET("/list", eventTopicList)
				eventTopic.POST("/add", addEventTopic)
				eventTopic.POST("/update", upEventTopic)
				eventTopic.POST("/delete", delEventTopic)
			}
			recommend := popular.Group("/recommend")
			{
				recommend.GET("/list", popRecommendList)
				recommend.POST("/add", addPopRecommend)
				recommend.POST("/update", upPopRecommend)
				recommend.POST("/delete", delPopRecommend)
			}
			stars := popular.Group("/stars")
			{
				stars.GET("/list", popularStarsList)
				stars.POST("/add", addPopularStars)
				stars.POST("/update", updatePopularStars)
				stars.POST("/delete", deletePopularStars)
				stars.POST("/reject", rejectPopularStars)
			}
			selected := popular.Group("/selected")
			{
				selected.GET("/series", selSeries)
				selected.POST("/sort", selSort)
				selected.GET("/export", selExport)
				selected.GET("/list", selList)
				selected.POST("/add", selAdd)
				selected.POST("/edit", selEdit)
				selected.GET("/arc_preview", arcPreview)
				selected.POST("/delete", selDelete)
				selected.POST("/reject", selReject)
				selected.GET("/preview", selPreview)
				selected.POST("/audit_serie", authSvc.Permit2("POPULAR_CARD_ADMIN"), selSerieAudit)
				selected.POST("/edit_serie", selSerieEdit)
				selected.POST("/touch_user", authSvc.Permit2("POPULAR_CARD_ADMIN"), selTouchUsers)
				selected.GET("/series_in_use", selSeriesInUse)
				selected.GET("/latestSelPreview", latestSelPreview) // 给创作中心用，返回最新一期每周必看核心信息
			}
			aggregation := popular.Group("/aggregation")
			{
				aggregation.GET("/list", aggregationList)
				aggregation.POST("/save", aggregationSave)
				aggregation.POST("/operate", aggOperate)
				aggregation.GET("/view", aggView)
				aggregation.POST("/view/add", aggViewAdd)
				aggregation.POST("/view/operate", aggViewOp)
				aggregation.GET("/tag", tag)
				aggregation.POST("/tag/add", aggTagAdd)
				aggregation.POST("/tag/del", aggTagDel)
			}
			aggregation2 := popular.Group("/aggregation/v2")
			{
				aggregation2.GET("/list", list)
				aggregation2.POST("/operate", operate)
				aggregation2.POST("/add", add)
				aggregation2.POST("/save", save)
				aggregation2.GET("/view", view)
				aggregation2.POST("/view/add", viewAdd)
				aggregation2.POST("/view/operate", viewOperate)
				aggregation2.POST("/tag/add", tagAdd)
				aggregation2.POST("/tag/del", tagDel)
			}
			entrance := popular.Group("/entrance") // 热门顶部入口配置
			{
				entrance.GET("/list", popularEntrance)
				entrance.POST("/save", popEntranceSave)
				entrance.POST("/operate", popEntranceOperate) // 启用，禁用，删除
				entrance.POST("/red_dot/update", vfySvr.Verify, redDotUpdate)
				entrance.POST("/red_dot/update_disposable", redDotUpdateDisposable) // 一次性红点后台更新红点信息
				entrance.GET("/view", popEntranceView)
				entrance.POST("/view/save", popEntranceViewSave)
				entrance.POST("/view/add", popEntranceViewAdd)
				entrance.POST("/view/operate", popEntranceViewOperate)
				entrance.POST("/tag/add", popEntranceTagAdd)
				entrance.POST("/tag/del", popEntranceTagDel)
				entrance.POST("/middle/save", popMiddleSave)
				entrance.GET("/middle/list", popMiddleList)
			}
			largeCard := popular.Group("/large/card") //  热点大卡配置
			{
				largeCard.GET("/list", popLargeCardList)
				largeCard.POST("/save", popLargeCardSave)
				largeCard.POST("/operate", popLargeCardOperate)
			}
			liveCard := popular.Group("/live/card") //  直播小卡配置
			{
				liveCard.GET("/list", popLiveCardList)
				liveCard.POST("/save", popLiveCardSave)
				liveCard.POST("/operate", popLiveCardOperate)
			}
			articleCard := popular.Group("/article/card") //  专栏小卡配置
			{
				articleCard.GET("/list", articleCardList)
				articleCard.POST("/save", articleCardSave)
				articleCard.POST("/operate", articleCardOperate)
			}
		}
		informationGroup := feed.Group("/information")
		{
			// 资讯卡片推荐配置
			recommend := informationGroup.Group("/recommend")
			{
				recommend.GET("/list", authSvc.Permit2("INFORMATION_RECOMMEND_CARD_EDIT"), recommendCardList)
				recommend.POST("/add", authSvc.Permit2("INFORMATION_RECOMMEND_CARD_EDIT"), addRecommendCard)
				recommend.POST("/modify", authSvc.Permit2("INFORMATION_RECOMMEND_CARD_EDIT"), modifyRecommendCard)
				recommend.POST("/delete", authSvc.Permit2("INFORMATION_RECOMMEND_CARD_EDIT"), deleteRecommendCard)
				recommend.POST("/offline", authSvc.Permit2("INFORMATION_RECOMMEND_CARD_EDIT"), offlineRecommendCard)
				recommend.POST("/pass", authSvc.Permit2("INFORMATION_RECOMMEND_CARD_ADMIN"), passRecommendCard)
				recommend.POST("/reject", authSvc.Permit2("INFORMATION_RECOMMEND_CARD_ADMIN"), rejectRecommendCard)
			}
		}

		auth := feed.Group("/auth")
		{
			rcmd := auth.Group("/rcmd")
			{
				rcmd.GET("/web", middleSvc.Permit("WEB_RCMD_ADMIN,WEB_RCMD_EDIT"), webRcmdAuth)
				rcmd.GET("/app", middleSvc.Permit("POS_REC_PLUS_APPLY,POS_REC_PLUS_CHECK"), appRcmdAuth)
			}
		}
		web := feed.Group("/web")
		{
			rcmd := web.Group("/rcmd")
			{
				rcmdCard := rcmd.Group("/card")
				{
					rcmdCard.GET("/list", webRcmdCardList)
					rcmdCard.POST("/add", addWebRcmdCard)
					rcmdCard.POST("/update", upWebRcmdCard)
					rcmdCard.POST("/delete", delWebRcmdCard)
				}

				//rcmdGroup := rcmd.Group("", middleSvc.Permit("WEB_RCMD_ADMIN,WEB_RCMD_EDIT"))
				//暂时修改web相关推荐旧api入口, 后续稳定可该模块代码删除
				rcmdGroup := rcmd.Group("/ban-api", middleSvc.Permit("WEB_RCMD_ADMIN,WEB_RCMD_EDIT"))
				{
					rcmdGroup.GET("/list", webRcmdList)
					rcmdGroup.POST("/add", addWebRcmd)
					rcmdGroup.POST("/update", upWebRcmd)
					rcmdGroup.POST("/delete", delWebRcmd)
					rcmdGroup.POST("/opt", optWebRcmd)
				}
			}
		}
		dynamic := feed.Group("/dynamic", authSvc.Permit2("DYNAMIC_SEARCH"))
		{
			search := dynamic.Group("/search")
			{
				search.GET("/list", dySeachList)
				search.POST("/add", addDySeach)
				search.POST("/update", updateDySeach)
				search.POST("/del", delDySeach)
			}
		}
		config := feed.Group("/config", authSvc.Permit2("CUSTOM_CONFIG"))
		{
			config.GET("", getConfig)
			config.GET("/list", configList)
			config.POST("/add", configAdd)
			config.POST("/edit", configEdit)
			config.POST("/opt", configOpt)
			config.GET("/log", configLog)
		}
		menu := feed.Group("/menu")
		{
			tab := menu.Group("/tab", authSvc.Permit2("MENU_TAB_EXT"))
			{
				tab.GET("", menuTabList)
				tab.POST("/save", menuTabSave)
				tab.POST("/operate", menuTabOperate)
				tab.GET("/search/navigation", menuSearch)
			}
			skin := menu.Group("/skin", authSvc.Permit2("MENU_SKIN_EXT"))
			{
				skin.GET("", skinList)
				skin.POST("/save", skinSave)
				skin.POST("/operate", skinEdit)
				skin.GET("/search", skinSearch)
			}
		}
		bubble := feed.Group("/bubble", authSvc.Permit2("BUBBLE_CONFIG"))
		{
			bubble.GET("/list", bubbleList)
			bubble.GET("/position/list", bubblePositionList)
			bubble.POST("/add", bubbleAdd)
			bubble.POST("/edit", bubbleEdit)
			bubble.POST("/state", bubbleState)
		}
		hideEntrance := feed.Group("/hide/entrance", authSvc.Permit2("HIDE_ENTRANCE"))
		{
			hideEntrance.GET("", hiddenList)
			hideEntrance.GET("/detail", hiddenDetail)
			hideEntrance.POST("/save", hiddenSave)
			hideEntrance.POST("/opt", hiddenOpt)
			hideEntrance.GET("/search", entranceSearch)
		}
		mogul := feed.Group("/mogul", authSvc.Permit2("MOGUL"))
		{
			mogul.GET("/applog/list", appMogulLogList)
		}
		icon := feed.Group("/icon", authSvc.Permit2("MNG_ICON"))
		{
			icon.GET("/list", iconList)
			icon.GET("/detail", iconDetail)
			icon.POST("/save", iconSave)
			icon.POST("/opt", iconOpt)
			icon.GET("/module", iconModule)
		}
		sidebar := feed.Group("/sidebar", authSvc.Permit2("SIDEBAR"))
		{
			sidebar.GET("/module/list", moduleList)
			sidebar.GET("/module/detail", moduleDetail)
			sidebar.POST("/module/save", moduleSave)
			sidebar.POST("/module/opt", moduleOpt)
			sidebar.GET("/module/item/list", moduleItemList)
			sidebar.GET("/module/item/detail", moduleItemDetail)
			sidebar.POST("/module/item/save", moduleItemSave)
			sidebar.POST("/module/item/opt", moduleItemOpt)
		}
		intervention := feed.Group("/intervention")
		{
			intervention.POST("/create", authSvc.Permit2("RCMD_LIST_INTER"), createIntervention)
			intervention.POST("/edit", authSvc.Permit2("RCMD_LIST_INTER"), editIntervention)
			intervention.POST("/changeStatus", authSvc.Permit2("RCMD_LIST_INTER"), changeIntervention)
			intervention.GET("/list", authSvc.Permit2("RCMD_LIST_INTER"), searchIntervention)
			intervention.GET("/optLogs", authSvc.Permit2("RCMD_LIST_INTER"), searchInterventionLogs)
			intervention.POST("/createOptLog", authSvc.Permit2("RCMD_LIST_INTER"), createOptLog)
		}
		tianma := feed.Group("/tianma")
		{
			tianma.GET("/boss/upload/signedUrl", bossSignedUploadUrl)        // 获取上传到boss的预签名url
			tianma.GET("/boss/download/signedUrl", bossSignedDownloadUrl)    // 通过key获取到下载用的预签名url，有有效期
			tianma.POST("/posRec/fileInfo/update", updateMidFileInfo)        // 更新某个推荐对应的文件信息
			tianma.GET("/common/isHdfsPathAccessible", IsHdfsPathAccessible) //判断hdfs路径是否可访问
			tianma.GET("/common/isHttpPathAccessible", IsHttpPathAccessible) //判断http路径是否可访问
			// 天马业务弹窗配置
			tianma.GET("/popup/list", authSvc.Permit2("TIANMA_POPUP_ADMIN"), popupConfigList)      // 获取弹窗配置列表
			tianma.POST("/popup/add", authSvc.Permit2("TIANMA_POPUP_ADMIN"), popupConfigAdd)       // 新增弹窗配置
			tianma.POST("/popup/edit", authSvc.Permit2("TIANMA_POPUP_ADMIN"), popupConfigEdit)     // 修改弹窗配置
			tianma.POST("/popup/delete", authSvc.Permit2("TIANMA_POPUP_ADMIN"), popupConfigDelete) // 删除弹窗配置
			tianma.POST("/popup/audit", authSvc.Permit2("TIANMA_POPUP_ADMIN"), popupConfigAudit)   // 审核弹窗配置（暂时只用来做下线）
		}
		splash := feed.Group("/splash", authSvc.Permit2("SPLASH")) // 品牌闪屏
		{
			// 闪屏物料
			splash.POST("/image/add", splashImageAdd)       // 新增物料
			splash.POST("/image/edit", splashImageEdit)     // 修改物料
			splash.POST("/image/delete", splashImageDelete) // 删除物料
			splash.GET("/image/list", splashImageList)      // 获取物料列表
			// 闪屏默认配置
			splash.POST("/config/audit", splashConfigAudit) // 审核配置
			splash.POST("/config/add", splashConfigAdd)     // 新建配置
			splash.POST("/config/edit", splashConfigEdit)   // 编辑配置
			splash.GET("/config/list", splashConfigList)    // 获取配置列表
			// 闪屏自选配置
			splash.POST("/config/select/audit", splashConfigSelectAudit)               // 审核配置
			splash.POST("/config/select/save", splashConfigSelectSave)                 // 新建配置
			splash.GET("/config/select/list", splashConfigSelectList)                  // 获取配置列表
			splash.POST("/config/select/sortBoundary", splashConfigSelectSortBoundary) // 置顶/置底配置
			splash.POST("/config/select/delete", splashConfigSelectDelete)             // 删除配置
			// 闪屏自选配置分类
			splash.GET("/category/all", splashCategoryListAll)
			splash.POST("/category/save", splashCategorySave)
		}
		// app首页顶部入口配置
		entry := feed.Group("/entry")
		{
			// 新增入口
			entry.POST("/create", authSvc.Permit2("APP_ENTRY_CONFIG"), createEntry)
			// 删除入口
			entry.POST("/delete", authSvc.Permit2("APP_ENTRY_CONFIG"), deleteEntry)
			// 编辑入口
			entry.POST("/edit", authSvc.Permit2("APP_ENTRY_CONFIG"), editEntry)
			// 入口上下线状态切换
			entry.POST("/toggle", authSvc.Permit2("APP_ENTRY_CONFIG"), toggleEntry)
			// 获取入口策略列表
			entry.GET("/list", authSvc.Permit2("APP_ENTRY_CONFIG"), getEntryList)
			// 设定新的状态切换时间
			entry.POST("/timeSettings/setNext", authSvc.Permit2("APP_ENTRY_CONFIG"), setNextTimeSettings)
			// 获取状态切换时间列表
			entry.GET("/timeSettings/list", authSvc.Permit2("APP_ENTRY_CONFIG"), getTimeSettingList)
			// 获取当前有效的状态
			// entry.GET("/current", authSvc.Permit2("APP_ENTRY_CONFIG"), getCurrentEntry)
		}
		feature := feed.Group("/feature")
		{
			app := feature.Group("/app", authSvc.Permit2("FEATURE_APP_ADMIN"))
			{
				app.GET("/list", appList)
				app.GET("/plat", appPlat)
				app.POST("/save", authSvc.Permit2("FEATURE_APP_EDIT"), saveApp)
			}
			build := feature.Group("/build", authSvc.Permit2("FEATURE_BUILD_LIMIT"))
			{
				build.GET("/list", buildList)
				build.POST("/save", saveBuild)
				build.POST("/act", handleBuild)
			}
			sw := feature.Group("/switch")
			{
				sw.GET("/tv/list", switchTvList)
				sw.POST("/tv/save", switchTvSave)
				sw.POST("/tv/del", switchTvDel)
			}
			businessConfig := feature.Group("/business", authSvc.Permit2("FEATURE_APP_ADMIN"))
			{
				businessConfig.GET("/list", businessConfigList)
				businessConfig.POST("/save", businessConfigSave)
				businessConfig.POST("/act", businessConfigAct)
			}
			abtest := feature.Group("/abtest") // authSvc.Permit2("FEATURE_ABTEST")
			{
				abtest.GET("/list", abtestList)
				abtest.POST("/save", abtestSave)
				abtest.POST("/act", abtestHandle)
			}
		}
		card := feed.Group("/card")
		{
			navigation := card.Group("/navigation", authSvc.Permit2("NAVIGATION_CARD"))
			{
				navigation.POST("/add", addNavigationCard)
				navigation.POST("/update", updateNavigationCard)
				navigation.POST("/delete", deleteNavigationCard)
				navigation.GET("/info", queryNavigationCard)
				navigation.GET("/list", listNavigationCard)
			}
			content := card.Group("/content", authSvc.Permit2("SEARCH_CONTENT_CARD"))
			{
				content.POST("/add", addContentCard)
				content.POST("/update", updateContentCard)
				content.POST("/delete", deleteContentCard)
				content.GET("/info", queryContentCard)
				content.GET("/list", listContentCard)
			}
		}
		frontpageG := feed.Group("/frontpage")
		{
			frontpageG.GET("/detail/:resource/:id", authSvc.Permit2("FRONT_PAGE_SEE"), getFrontpageDetail)
			frontpageG.GET("/list", authSvc.Permit2("FRONT_PAGE_SEE"), listFrontpages)
			frontpageG.GET("/menus", authSvc.Permit2("FRONT_PAGE_SEE"), listFrontpageMenus)
			frontpageG.POST("/add", authSvc.Permit2("FRONT_PAGE_MANAGER"), addFrontpage)
			//frontpageG.POST("/batchAdd", authSvc.Permit2("FRONT_PAGE_MANAGER"), batchAddFrontpages)
			frontpageG.POST("/edit", authSvc.Permit2("FRONT_PAGE_MANAGER"), editFrontpage)
			frontpageG.POST("/action", authSvc.Permit2("FRONT_PAGE_MANAGER"), actionFrontpage)
			frontpageG.GET("/history", authSvc.Permit2("FRONT_PAGE_LOG"), listFrontpageHistory)
			frontpageG.GET("/location/policyList", authSvc.Permit2("FRONT_PAGE_MANAGER"), listFrontpageLocationPolicies)
		}
		appealGroup := feed.Group("/pwd_appeal", authSvc.Permit2("PWD_APPEAL"))
		{
			appealGroup.GET("/list", pwdAppealList)
			appealGroup.GET("/photo", pwdAppealPhoto)
			appealGroup.POST("/audit/pass", passPwdAppeal)
			appealGroup.POST("/audit/reject", rejectPwdAppeal)
			appealGroup.GET("/export", exportPwdAppeal)
		}
		spmodeGroup := feed.Group("/special_mode", authSvc.Permit2("TEENAGER_PASSWORD"))
		{
			spmodeGroup.GET("/search", searchSpmode)
			spmodeGroup.POST("/relieve", relieveSpmode)
			spmodeGroup.GET("/log", spmodeLog)
		}
		notice := feed.Group("/package/push", authSvc.Permit2("NOTICE"))
		{
			notice.GET("/list", pushList)
			notice.POST("/save", pushSave)
			notice.GET("/detail", pushDetail)
			notice.POST("/delete", pushDelete)
		}
		antiCrawlerGroup := feed.Group("/anti_crawler", authSvc.Permit2("ANTI_CRAWLER"))
		{
			antiCrawlerGroup.GET("/user_log", antiCrawlerUserLog)
			businessConfig := antiCrawlerGroup.Group("/business")
			{
				businessConfig.GET("/list", antiCrawlerBusinessConfigList)
				businessConfig.POST("/update", antiCrawlerBusinessConfigUpdate)
				businessConfig.POST("/delete", antiCrawlerBusinessConfigDelete)
			}
		}
		teenManualGroup := feed.Group("/teen_manual", authSvc.Permit2("TEENAGER_MANUAL_FORCE"))
		{
			teenManualGroup.GET("/search", searchTeenManual)
			teenManualGroup.POST("/open", authSvc.Permit2("TEENAGER_MANUAL_OPEN"), openTeenManual)
			teenManualGroup.POST("/quit", authSvc.Permit2("TEENAGER_MANUAL_QUIT"), quitTeenManual)
			teenManualGroup.GET("/log", teenManualLog)
		}
		familyGroup := feed.Group("/family", authSvc.Permit2("FAMILY"))
		{
			familyGroup.GET("/search", searchFamily)
			familyGroup.GET("/bind/list", familyBindList)
			familyGroup.POST("/unbind", authSvc.Permit2("FAMILY_UNBIND"), unbindFamily)
		}
	}
}

// ping check server ok.
func ping(c *bm.Context) {

}
