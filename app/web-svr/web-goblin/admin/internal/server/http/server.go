package http

import (
	"net/http"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/web-svr/web-goblin/admin/internal/service"
)

var (
	svc *service.Service
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
	var (
		hc struct {
			Server *bm.ServerConfig
			Auth   *warden.ClientConfig
		}
	)
	if err = paladin.Get("http.toml").UnmarshalTOML(&hc); err != nil {
		if err != paladin.ErrNotExist {
			return
		}
		err = nil
	}
	svc = s
	engine = bm.DefaultServer(hc.Server)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/admin/goblin")
	{
		bGroup := g.Group("/business")
		{
			bGroup.GET("/info", infoBusiness)
			bGroup.GET("/list", listBusiness)
			bGroup.POST("/add", addBusiness)
			bGroup.POST("/save", editBusiness)
			bGroup.POST("/del", delBusiness)
		}
		cGroup := g.Group("/customer")
		{
			cGroup.GET("/info", infoCustomer)
			cGroup.GET("/list", listCustomer)
			cGroup.POST("/add", addCustomer)
			cGroup.POST("/save", editCustomer)
			cGroup.POST("del", delCustomer)
		}
	}
}

func ping(ctx *bm.Context) {
	if err := svc.Ping(ctx); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
