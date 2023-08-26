package http

import (
	"net/http"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/service"
)

var svc *service.Service

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
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/admin/gateway-dev-management")
	{
		eg := g.Group("/expression")
		{
			eg.POST("/device", checkExpressionWithDevice)
			eg.POST("/context", checkExpressionWithContext)
		}
		rg := g.Group("/railgun")
		{
			rg.POST("/bot", svc.RailgunBotHandler())
		}
		mg := g.Group("/monitor")
		{
			mg.POST("/rule/config", thresholdConfig)
			mg.GET("/rule/team", customizedRules)
			mg.POST("/rule/edit", editRule)
			mg.POST("/rule/delete", deleteRule)
			mg.GET("/tree", fetchRoleTree)
			mg.POST("/receiver/root", rootReceiverGroup)
			mg.GET("/receiver/my", myService)
		}
		bg := g.Group("/bot")
		{
			bg.GET("", botVerify)
			bg.POST("", botCallback)
			bg.GET("/deploy", deployment)
			bg.GET("/getDeploy", getDeploy)
			bg.GET("/start", startDeploy)
			bg.GET("/resume", resumeDeploy)
			bg.GET("/done", doneDeploy)
			bg.GET("/rollback", rollbackDeploy)
			bg.GET("/script", scripts)
			bg.POST("/newScript", newScript)
			bg.GET("/getScript", getScript)
			bg.GET("/doScript", doScript)
		}
	}
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
