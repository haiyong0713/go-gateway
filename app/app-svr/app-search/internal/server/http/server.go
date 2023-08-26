package http

import (
	abtest "go-common/component/tinker/middleware/http"
	"go-common/library/conf/env"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/cache"
	"go-common/library/net/http/blademaster/middleware/cache/store"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-interface/interface-legacy/middleware/stat"
	"go-gateway/app/app-svr/app-search/configs"
	service "go-gateway/app/app-svr/app-search/internal/service/v1"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	"net/http"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-search/api/v1"
)

var (
	srcSvr  *service.Service
	authSvc *auth.Auth
	// cache components
	hotSearchCacheSvr *cache.Cache
	newSearchCacheSvr *cache.Cache
	deg               *cache.Degrader
	// feature service
	featureSvc *feature.Feature
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
	var (
		cfg         bm.ServerConfig
		ct          paladin.TOML
		quotaConfig quota.Config
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	if err = ct.Get("Quota").UnmarshalTOML(&quotaConfig); err != nil {
		return
	}
	srcSvr = s
	engine = bm.DefaultServer(&cfg)
	authSvc = auth.New(nil)
	limiter := quota.New(&quotaConfig)
	hotSearchCacheSvr = cache.New(store.NewMemcache(configs.MemcacheDegradeConfig.Memcache))
	if env.Zone == "sh004" {
		newSearchCacheSvr = cache.New(store.NewMemcache(configs.MemcacheDegradeConfig.NewSearchMemcacheJd))
	} else {
		newSearchCacheSvr = cache.New(store.NewMemcache(configs.MemcacheDegradeConfig.NewSearchMemcacheYlf))
	}
	deg = cache.NewDegrader(configs.MemcacheDegradeConfig.Expire)
	engine.Use(limiter.Handler())
	pb.RegisterSearchBMServer(engine, srcSvr)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(anticrawler.Report())
	search := e.Group("/x/v2/search", featureSvc.BuildLimitHttp())
	search.Use(bm.CORS())
	search.GET("", authSvc.GuestMobile, abtest.Handler(), newSearchDegradeHandler(stat.SearchDegreeArgs...), arcmid.BatchPlayArgs(), searchAll)
	search.GET("/type", authSvc.GuestMobile, abtest.Handler(), newSearchDegradeHandler(stat.SearchTypeDegreeArgs...), searchByType)
	search.GET("/converge", authSvc.GuestMobile, searchConverge)
	search.GET("/episodes", authSvc.GuestMobile, searchEpisodes)
	search.GET("/episodes_new", authSvc.GuestMobile, searchEpisodesNew)
	search.GET("/live", authSvc.GuestMobile, searchLive)
	search.GET("/hot", authSvc.GuestMobile, hotSearchDegradeHandler(stat.HotSearchDegreeArgs...), hotSearch)
	search.GET("/trending", authSvc.GuestMobile, hotSearchDegradeHandler(stat.HotSearchDegreeArgs...), trending)
	search.GET("/trending/ranking", authSvc.Guest, hotSearchDegradeHandler(stat.HotSearchDegreeArgs...), ranking)
	search.GET("/suggest", authSvc.GuestMobile, suggest)
	search.GET("/suggest2", authSvc.GuestMobile, suggest2)
	search.GET("/suggest3", authSvc.GuestMobile, suggest3)
	search.GET("/defaultwords", authSvc.GuestMobile, defaultWords)
	search.GET("/user", authSvc.GuestMobile, searchUser)
	search.GET("/recommend", authSvc.GuestMobile, recommend)
	search.GET("/recommend/noresult", authSvc.GuestMobile, recommendNoResult)
	search.GET("/recommend/pre", authSvc.GuestMobile, recommendPre)
	search.GET("/recommend/tags", authSvc.GuestMobile, recommendTags)
	search.GET("/resource", authSvc.GuestMobile, resource)
	search.GET("/channel", authSvc.GuestMobile, searchChannel)
	search.GET("/square", authSvc.GuestMobile, hotSearchDegradeHandler(stat.SquareSearchDegreeArgs...), searchSquare)
	search.GET("/channel2", authSvc.GuestMobile, arcmid.BatchPlayArgs(), searchChannel2)
	search.GET("/siri/resolve/command", authSvc.GuestMobile, searchSiri)
}

func ping(ctx *bm.Context) {
	if _, err := srcSvr.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
