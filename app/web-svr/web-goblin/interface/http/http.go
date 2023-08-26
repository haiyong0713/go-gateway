package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/web-svr/web-goblin/interface/conf"
	"go-gateway/app/web-svr/web-goblin/interface/service/thirdsdk"
	"go-gateway/app/web-svr/web-goblin/interface/service/web"
	"go-gateway/app/web-svr/web-goblin/interface/service/wechat"
)

var (
	srvWeb      *web.Service
	srvWechat   *wechat.Service
	svrThirdsdk *thirdsdk.Service
	authSvr     *auth.Auth
)

// Init init .
func Init(c *conf.Config) {
	authSvr = auth.New(c.Auth)
	srvWeb = web.New(c)
	srvWechat = wechat.New(c)
	svrThirdsdk = thirdsdk.New(c)
	engine := bm.DefaultServer(c.BM)
	router(engine)
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

func router(e *bm.Engine) {
	e.Ping(ping)
	e.Register(register)
	e.Use(anticrawler.Report())
	group := e.Group("/x/web-goblin")
	{
		miGroup := group.Group("/mi")
		{
			miGroup.GET("/full", fullshort)
		}
		channelGroup := group.Group("/channel")
		{
			channelGroup.GET("", authSvr.Guest, channel)
		}
		ugcGroup := group.Group("ugc")
		{
			ugcGroup.GET("/full", ugcfull)
			ugcGroup.GET("/increment", ugcincre)
			ugcGroup.GET("/rank", ranking)
		}
		pgcGroup := group.Group("pgc")
		{
			pgcGroup.GET("/full", pgcfull)
			pgcGroup.GET("/increment", pgcincre)
		}
		weChatGroup := group.Group("/wechat")
		{
			weChatGroup.GET("/qrcode", qrcode)
			weChatGroup.POST("/push", push)
		}
		hisGroup := group.Group("/history")
		{
			hisGroup.GET("/search", authSvr.User, hisSearch)
		}
		thridsdk := group.Group("/thirdsdk")
		{
			thridsdk.GET("/author/bind/state", authSvr.User, authorBindState)
		}
		//group.GET("/share/encourage", authSvr.User, encourage)
		group.GET("/recruit", recruit)
		group.GET("/customer/center", cusCenter)
		group.GET("/out/arc", outArc)
		group.GET("/baidu/pusharc/content", baiduPushArcContent)
	}
}

func ping(c *bm.Context) {
	if err := srvWeb.Ping(c); err != nil {
		log.Error("web-goblin ping error(%v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)
}
