package http

import (
	"net/http"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/verify"
	pb "go-gateway/app/app-svr/collection-splash/api"
	"go-gateway/app/app-svr/collection-splash/internal/service"
)

var (
	svc    *service.Service
	vfySvc *verify.Verify
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
	vfySvc = verify.New(nil)
	pb.RegisterCollectionSplashBMServer(engine, s)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/internal/collection-splash", vfySvc.Verify)
	{
		g.GET("", splash)
		g.GET("/list", splashList)
		g.POST("/add", addSplash)
		g.POST("/update", updateSplash)
		g.POST("/delete", deleteSplash)
	}
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
