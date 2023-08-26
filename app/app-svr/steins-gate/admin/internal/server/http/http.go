package http

import (
	"go-gateway/app/app-svr/steins-gate/admin/conf"
	"go-gateway/app/app-svr/steins-gate/admin/internal/service"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/permit"
)

var (
	svc *service.Service
	//nolint:unused
	authSvr   *auth.Auth
	permitSvc = permit.New2(nil)
)

// New new a bm server.
func New(c *conf.Config, s *service.Service) (engine *bm.Engine) {
	authSvr = auth.New(nil)
	svc = s
	engine = bm.DefaultServer(c.Server)
	initRouter(engine)
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
		g.GET("/graph/show", permitSvc.Permit2("STEINS_SHOW"), graphShow)
		g.GET("/nodeinfo/audit", permitSvc.Permit2("STEINS_SHOW"), nodeInfoAudit)
		g.GET("/edgeinfo_v2/audit", permitSvc.Permit2("STEINS_SHOW"), edgeInfoV2Audit)
	}
}

func ping(c *bm.Context) {}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)

}
