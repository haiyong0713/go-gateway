package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"

	"go-gateway/app/app-svr/steins-gate/service/conf"
	"go-gateway/app/app-svr/steins-gate/service/internal/service"
)

var (
	svc     *service.Service
	authSvr *auth.Auth
	idfSvc  *verify.Verify
)

// New new a bm server.
func New(c *conf.Config, s *service.Service) (engine *bm.Engine) {
	authSvr = auth.New(nil)
	idfSvc = verify.New(nil)
	svc = s
	engine = bm.DefaultServer(c.Server)
	initRouter(engine)
	innerRouter(engine)
	if err := engine.Start(); err != nil {
		panic(err)
	}
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Register(register)
	g := e.Group("/x/stein")
	{
		g.POST("/graph/save", authSvr.User, saveGraph)
		g.POST("/mark", authSvr.User, mark)
		g.GET("/graph/show", authSvr.User, graphShow)
		g.GET("/graph/latest/list", authSvr.User, latestGraphList)
		g.GET("/graph/msg/check", authSvr.User, msgCheck)
		g.GET("/playurl", authSvr.User, playurl)
		g.GET("/graph/check", authSvr.User, graphCheck)
		g.GET("/video_info", authSvr.User, videoInfo)
		g.GET("/nodeinfo/preview", authSvr.User, nodeinfoPreview)
		g.GET("/nodeinfo", authSvr.Guest, checkSignErr, nodeinfo)
		g.GET("/edgeinfo_v2", authSvr.Guest, checkSignErr, edgeinfoV2)
		g.GET("/edgeinfo_v2/preview", authSvr.Guest, checkSignErr, edgeV2infoPreview)
		g.GET("/manager", authSvr.User, managerGraph)
		g.GET("/recent_arcs", recentArcs)
		g.GET("/skin/list", authSvr.User, skinList)

		g.GET("/user/rank/list", authSvr.Guest, rankList)
		g.POST("/user/rank/score/submit", authSvr.User, rankScoreSubmit)
	}
}

func innerRouter(e *bm.Engine) {
	g := e.Group("/x/internal/stein")
	{
		g.POST("/graph/audit", idfSvc.Verify, graphAudit)
	}
}

func checkSignErr(ctx *bm.Context) {
	if mobiApp := ctx.Request.Form.Get("mobi_app"); mobiApp == "iphone" || mobiApp == "android" {
		idfSvc.Verify(ctx)
	}
}

func ping(ctx *bm.Context) {
	if err := svc.Ping(ctx); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)

}
