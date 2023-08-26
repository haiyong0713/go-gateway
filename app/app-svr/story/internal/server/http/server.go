package http

import (
	"net/http"

	abtest "go-common/component/tinker/middleware/http"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	pb "go-gateway/app/app-svr/story/api"
	"go-gateway/app/app-svr/story/internal/service"
)

var (
	svc     *service.Service
	authSvc *auth.Auth
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svc = s
	engine = bm.DefaultServer(&cfg)
	pb.RegisterStoryBMServer(engine, s)
	authSvc = auth.New(nil)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Use(anticrawler.Report())
	g := e.Group("/x/v2/feed")
	{
		g.GET("/index/story", authSvc.GuestMobile, arcmid.BatchPlayArgs(), abtest.Handler(), feedStory)
		g.GET("/index/space/story", authSvc.GuestMobile, arcmid.BatchPlayArgs(), spaceStory)
		g.GET("/index/space/story/cursor", authSvc.GuestMobile, arcmid.BatchPlayArgs(), spaceStoryCursor)
		g.GET("/index/dynamic/story", authSvc.GuestMobile, arcmid.BatchPlayArgs(), dynamicStory)
		g.GET("/index/story/cart", authSvc.GuestMobile, storyCart)
		g.GET("/index/story/game/status", authSvc.GuestMobile, storyGameStatus)
	}
	story := e.Group("/x/story")
	{
		story.GET("/index/story", authSvc.GuestMobile, arcmid.BatchPlayArgs(), abtest.Handler(), feedStory)
		story.GET("/index/space/story", authSvc.GuestMobile, arcmid.BatchPlayArgs(), spaceStory)
		story.GET("/index/space/story/cursor", authSvc.GuestMobile, arcmid.BatchPlayArgs(), spaceStoryCursor)
		story.GET("/index/dynamic/story", authSvc.GuestMobile, arcmid.BatchPlayArgs(), dynamicStory)
		story.GET("/index/story/cart", authSvc.GuestMobile, storyCart)
		story.GET("/index/story/game/status", authSvc.GuestMobile, storyGameStatus)
	}
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
