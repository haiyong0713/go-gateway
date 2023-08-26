package http

import (
	"io/ioutil"
	"net/http"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/verify"
	appmid "go-gateway/app/app-svr/app-resource/interface/http/middleware"
	"go-gateway/app/app-svr/misaka/interface/internal/service"
)

var (
	svc *service.Service
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine) {
	var (
		hc struct {
			Server *bm.ServerConfig
		}
	)
	if err := paladin.Get("http.toml").UnmarshalTOML(&hc); err != nil {
		if err != paladin.ErrNotExist {
			panic(err)
		}
	}
	svc = s
	engine = bm.DefaultServer(hc.Server)
	initRouter(engine, verify.New(nil))
	if err := engine.Start(); err != nil {
		panic(err)
	}
	return
}

func initRouter(e *bm.Engine, v *verify.Verify) {
	e.Ping(ping)
	e.Register(register)
	e.Use(bm.CORS())
	g := e.Group("/misaka", appmid.InjectTimestamp())
	{
		g.POST("/report", appReport)
		g.POST("/web/report", webReport)
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

func appReport(c *bm.Context) {
	var (
		body []byte
		err  error
		code int
	)
	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	code, _ = svc.AppReport(c, body)
	c.Status(code)
	return
}

func webReport(c *bm.Context) {
	var (
		body []byte
		err  error
		code int
	)
	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	code, _ = svc.WebReport(c, body)
	c.Status(code)
	return
}
