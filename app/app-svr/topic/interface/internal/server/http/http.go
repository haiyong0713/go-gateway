package http

import (
	"net/http"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	pb "go-gateway/app/app-svr/topic/interface/api"
	"go-gateway/app/app-svr/topic/interface/internal/service"
)

var (
	topicSvc  *service.Service
	authSvc   *auth.Auth
	verifySvc *verify.Verify
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
	topicSvc = s
	engine = bm.DefaultServer(&cfg)
	authSvc = auth.New(nil)
	verifySvc = verify.New(nil)
	limiter := quota.New(&quotaConfig)
	engine.Use(limiter.Handler())
	pb.RegisterTopicBMServer(engine, s)
	initRouter(engine)
	initInternalRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Use(bm.CORS(), anticrawler.Report())
	e.Ping(ping)
	root := e.Group("/x/topic")
	{
		// 话题发布场景
		pub := root.Group("/pub")
		pub.GET("/search", authSvc.Guest, searchPubTopics)
		pub.GET("/rcmd/search", authSvc.Guest, searchRcmdPubTopics)
		pub.GET("/mine", authSvc.User, usrPubTopics)
		pub.GET("/is_existed", authSvc.User, isAlreadyExistedTopic)
		pub.GET("/endpoint", verifySvc.Verify, authSvc.User, pubTopicEndpoint)
		pub.POST("/upload", verifySvc.Verify, authSvc.User, pubUpload)
		pub.GET("/events", authSvc.Guest, topicPubEvents)
		// 创建动作场景
		create := root.Group("/create", verifySvc.Verify, authSvc.User)
		create.GET("/jurisdiction", hasCreateJurisdiction)
		create.POST("/submit", createTopic)
		// 话题收藏场景
		fav := root.Group("/fav", authSvc.User)
		fav.GET("/sub/list", subFavTopics)
		fav.POST("/sub/add", addFav)
		fav.POST("/sub/cancel", cancelFav)
		// web提供接口
		web := root.Group("/web")
		web.GET("/details/top", authSvc.Guest, webTopicInfo)
		web.GET("/details/cards", authSvc.Guest, arcmid.BatchPlayArgs(), webTopicCards)
		web.GET("/details/fold", authSvc.Guest, arcmid.BatchPlayArgs(), webTopicFoldCards)
		web.GET("/fav/list", authSvc.User, webSubFavTopics)
		web.GET("/pub/endpoint", authSvc.User, pubTopicEndpoint)
		web.POST("/submit", authSvc.User, webCreateTopic)
		web.GET("/dynamic/rcmd", authSvc.Guest, webDynamicRcmdTopics)
		web.POST("/pub/upload", authSvc.User, pubUpload)
		// 社区场景
		community := root.Group("/community")
		community.GET("/hotword/videos", authSvc.Guest, arcmid.BatchPlayArgs(), hotWordVideos)
		community.GET("/hotword/dynamics", authSvc.Guest, arcmid.BatchPlayArgs(), hotWordDynamics)
		// 垂直通用场景
		vert := root.Group("/vert")
		vert.GET("/search", authSvc.Guest, vertSearchTopics)
		vert.GET("/center", authSvc.Guest, vertTopicCenter)
		vert.GET("/online", authSvc.Guest, vertTopicOnline)
		general := root.Group("/general")
		general.GET("/feed/list", authSvc.Guest, arcmid.BatchPlayArgs(), generalFeedList)
		// 其他场景
		root.POST("/report", authSvc.User, topicReport)
		root.POST("/resource/report", authSvc.User, topicResReport)
		root.POST("/like", authSvc.User, topicLike)
		root.POST("/dislike", authSvc.User, topicDislike)
		root.GET("/timeline", authSvc.Guest, topicTimeLine)
	}
}

func initInternalRouter(e *bm.Engine) {
	root := e.Group("/x/internal/topic/web", authSvc.Guest)
	{
		root.GET("/details/top", webTopicInfo)
		root.GET("/details/cards", arcmid.BatchPlayArgs(), webTopicCards)
		root.GET("/details/fold", arcmid.BatchPlayArgs(), webTopicFoldCards)
	}
}

func ping(ctx *bm.Context) {
	if _, err := topicSvc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
