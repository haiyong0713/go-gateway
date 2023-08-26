package http

import (
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/proxy"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/http/blademaster/render"

	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/service/act"
	"go-gateway/app/app-svr/app-show/interface/service/banner"
	"go-gateway/app/app-svr/app-show/interface/service/daily"
	pingSvr "go-gateway/app/app-svr/app-show/interface/service/ping"
	"go-gateway/app/app-svr/app-show/interface/service/rank"
	ranklist "go-gateway/app/app-svr/app-show/interface/service/rank-list"
	"go-gateway/app/app-svr/app-show/interface/service/region"
	"go-gateway/app/app-svr/app-show/interface/service/show"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

var (
	// depend service
	authSvc *auth.Auth
	// self service
	bannerSvc   *banner.Service
	regionSvc   *region.Service
	showSvc     *show.Service
	pingSvc     *pingSvr.Service
	rankSvc     *rank.Service
	dailySvc    *daily.Service
	actSvc      *act.Service
	rankListSvc *ranklist.Service
	verifySvc   *verify.Verify
	featureSvc  *feature.Feature
	cfg         *conf.Config
)

func Init(c *conf.Config, svr *Server) {
	initService(c, svr)
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
func initService(c *conf.Config, svr *Server) {
	authSvc = svr.AuthSvc
	bannerSvc = svr.BannerSvc
	regionSvc = svr.RegionSvc
	showSvc = svr.ShowSvc
	pingSvc = svr.PingSvc
	rankSvc = svr.RankSvc
	dailySvc = svr.DailySvc
	actSvc = svr.ActSvr
	rankListSvc = svr.RankListSvc
	verifySvc = svr.VerifySvr
	featureSvc = svr.FeatureSvr
	cfg = c
}

type Server struct {
	// depend service
	AuthSvc *auth.Auth
	// self service
	BannerSvc   *banner.Service
	RegionSvc   *region.Service
	ShowSvc     *show.Service
	PingSvc     *pingSvr.Service
	RankSvc     *rank.Service
	DailySvc    *daily.Service
	ActSvr      *act.Service
	RankListSvc *ranklist.Service
	VerifySvr   *verify.Verify
	Config      *conf.Config
	FeatureSvr  *feature.Feature
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(bm.CORS())
	proxyHandler := proxy.NewZoneProxy("sh004", "http://sh001-app.bilibili.com")
	bnnr := e.Group("/x/v2/banner", authSvc.GuestMobile)
	{
		bnnr.GET("", banners)
	}
	activity := e.Group("/x/v2/activity", verifySvc.Verify, featureSvc.BuildLimitHttp())
	{
		activity.GET("/index", authSvc.GuestMobile, arcmid.BatchPlayArgs(), ActIndex)
		activity.GET("/inline", authSvc.GuestMobile, arcmid.BatchPlayArgs(), inlineTab)
		activity.GET("/menu", authSvc.GuestMobile, arcmid.BatchPlayArgs(), menuTab)
		activity.POST("/liked", authSvc.UserMobile, ActLiked)
		activity.GET("/detail", ActDetail)
		activity.GET("/like/list", authSvc.GuestMobile, arcmid.BatchPlayArgs(), LikeList)
		activity.GET("/supernatant", authSvc.GuestMobile, supernatant)
		activity.POST("/follow", proxyHandler, authSvc.UserMobile, actFollow)
		activity.GET("/base", baseDetail)
		activity.GET("/tab", actTab)
		activity.POST("/receive", proxyHandler, authSvc.UserMobile, actReceive)
	}
	region := e.Group("/x/v2/region", authSvc.GuestMobile, featureSvc.BuildLimitHttp())
	{
		region.GET("", regions)
		region.GET("/list", regionsList)
		region.GET("/index", regionsIndex)
		region.GET("/show", regionShow)
		region.GET("/show/dynamic", regionShowDynamic)
		region.GET("/show/child", regionChildShow)
		region.GET("/show/child/list", regionChildListShow)
		region.GET("/dynamic", regionDynamic)
		region.GET("/dynamic/list", regionDynamicList)
		region.GET("/dynamic/child", regionDynamicChild)
		region.GET("/dynamic/child/list", regionDynamicChildList)
	}
	rank := e.Group("/x/v2/rank", authSvc.GuestMobile)
	{
		rank.GET("", rankAll)
		rank.GET("/region", rankRegion)
	}
	rankList := e.Group("/x/v2/rank-list") // 榜单
	{
		rankList.GET("/index", authSvc.Guest, rankListIndex)
	}
	show := e.Group("/x/v2/show", authSvc.Guest, featureSvc.BuildLimitHttp())
	{
		show.GET("", shows)
		show.GET("/region", showsRegion)
		show.GET("/index", showsIndex)
		show.GET("/widget", showWidget)
		show.GET("/temp", showTemps)
		show.GET("/change", showChange)
		show.GET("/change/live", showLiveChange)
		show.GET("/change/region", showRegionChange)
		show.GET("/change/bangumi", showBangumiChange)
		show.GET("/change/dislike", showDislike)
		show.GET("/change/article", showArticleChange)
		show.GET("/popular", popular)
		show.GET("/popular/archive", popularArchive)
		show.GET("/popular/index", arcmid.BatchPlayArgs(), popular2)
		show.GET("/popular/good_history", precious)
		show.POST("/popular/good_history/sub/add", preciousSubAdd)
		show.POST("/popular/good_history/sub/del", preciousSubDel)
		show.GET("/popular/aggregation", aggregation)
		selected := show.Group("/popular/selected")
		{
			selected.GET("", proxyHandler, selectedSerie)
			selected.GET("/series", proxyHandler, series)
			selected.POST("/sub/add", addFav)
			selected.POST("/sub/del", delFav)
			selected.GET("/sub/status", checkFav)
		}
	}
	daily := e.Group("/x/v2/daily", authSvc.GuestMobile)
	{
		daily.GET("/list", dailyID)
	}
	column := e.Group("/x/v2/column", authSvc.GuestMobile)
	{
		column.GET("", columnList)
	}
	cg := e.Group("/x/v2/category", authSvc.GuestMobile)
	{
		cg.GET("", category)
	}
}

// returnJSON return json no message
func returnJSON(c *bm.Context, data interface{}, err error) {
	code := http.StatusOK
	c.Error = err
	bcode := ecode.Cause(err)
	c.Render(code, render.JSON{
		Code:    bcode.Code(),
		Message: "",
		Data:    data,
	})
}

// returnDataJSON return json no message
// nolint:unparam
func returnDataJSON(c *bm.Context, data map[string]interface{}, ttl int, err error) {
	code := http.StatusOK
	if ttl < 1 {
		ttl = 1
	}
	if err != nil {
		c.JSON(nil, err)
	} else {
		if data != nil {
			data["code"] = 0
			data["message"] = ""
			data["ttl"] = ttl
		}
		c.Render(code, render.MapJSON(data))
	}
}

func Close() {
	rankSvc.Close()
}
