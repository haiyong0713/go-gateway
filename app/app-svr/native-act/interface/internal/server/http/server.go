package http

import (
	"net/http"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-common/library/net/http/blademaster/middleware/auth"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/service"
	webdevmid "go-gateway/app/app-svr/native-act/interface/middleware/webdevice"
)

var (
	svc     *service.Service
	authSvc *auth.Auth
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
	engine, err = newEngine(s)
	if err != nil {
		return
	}
	initRouter(engine)
	err = engine.Start()
	return
}

func newEngine(s *service.Service) (*bm.Engine, error) {
	var (
		serverCfg bm.ServerConfig
		authCfg   *auth.Config
		ct        paladin.TOML
	)
	if err := paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return nil, err
	}
	if err := ct.Get("Server").UnmarshalTOML(&serverCfg); err != nil {
		return nil, err
	}
	if ct.Exist("Auth") {
		if err := ct.Get("Auth").UnmarshalTOML(authCfg); err != nil {
			return nil, err
		}
	}
	svc = s
	authSvc = auth.New(authCfg)
	return bm.DefaultServer(&serverCfg), nil
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	natGroup := e.Group("/x/native_act", webdevmid.BindWebdevice())
	{
		natGroup.GET("/index", authSvc.Guest, index)
		natGroup.GET("/dynamic", authSvc.Guest, dynamic)
		natGroup.GET("/editor", authSvc.Guest, editor)
		natGroup.GET("/resource", authSvc.Guest, resource)
		natGroup.GET("/video", authSvc.Guest, video)
		natGroup.POST("/vote", authSvc.User, vote)
		natGroup.POST("/reserve", authSvc.User, reserve)
		natGroup.GET("/supernatant/timeline", authSvc.Guest, timelineSupernatant)
		natGroup.GET("/supernatant/ogv", authSvc.Guest, ogvSupernatant)
		natGroup.POST("/follow_ogv", authSvc.User, followOgv)
		natGroup.GET("/progress", authSvc.Guest, progress)
		natGroup.GET("/bottom_tab", authSvc.Guest, bottomTab)
		natGroup.POST("/handle_click", authSvc.User, handleClick)
	}
}

func index(c *bm.Context) {
	req := new(api.IndexReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.Index(c, req))
}

func dynamic(c *bm.Context) {
	req := new(api.DynamicReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.Dynamic(c, req))
}

func editor(c *bm.Context) {
	req := new(api.EditorReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.Editor(c, req))
}

func resource(c *bm.Context) {
	req := new(api.ResourceReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.Resource(c, req))
}

func video(c *bm.Context) {
	req := new(api.VideoReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.Video(c, req))
}

func vote(c *bm.Context) {
	req := new(api.VoteReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.Vote(c, req))
}

func reserve(c *bm.Context) {
	req := new(api.ReserveReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.Reserve(c, req))
}

func timelineSupernatant(c *bm.Context) {
	req := new(api.TimelineSupernatantReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.TimelineSupernatant(c, req))
}

func ogvSupernatant(c *bm.Context) {
	req := new(api.OgvSupernatantReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.OgvSupernatant(c, req))
}

func followOgv(c *bm.Context) {
	req := new(api.FollowOgvReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.FollowOgv(c, req))
}

func progress(c *bm.Context) {
	req := new(api.ProgressReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.Progress(c, req))
}

func bottomTab(c *bm.Context) {
	req := new(api.BottomTabReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.BottomTab(c, req))
}

func handleClick(c *bm.Context) {
	req := new(api.HandleClickReq)
	if err := c.BindWith(req, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(svc.HandleClick(c, req))
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
