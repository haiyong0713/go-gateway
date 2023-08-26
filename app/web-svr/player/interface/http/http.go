package http

import (
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/proxy"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/player/interface/conf"
	"go-gateway/app/web-svr/player/interface/service"
	"go-gateway/pkg/idsafe/bvid"
)

var (
	authSvr *auth.Auth
	vfySvr  *verify.Verify
	playSvr *service.Service
)

// Init init http.
func Init(c *conf.Config, s *service.Service) {
	authSvr = auth.New(c.Auth)
	vfySvr = verify.New(c.Verify)
	playSvr = s
	engine := bm.NewServer(c.BM.Outer)
	engine.Use(bm.Recovery(), bm.Trace(), bm.Logger(), bm.Mobile())
	outerRouter(engine)
	internalRouter(engine)
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

func outerRouter(e *bm.Engine) {
	e.Use(bm.CORS(), bm.CSRF())
	e.GET("/monitor/ping", ping)
	e.GET("/x/player.so", authSvr.Guest, player)
	proxyHandler := proxy.NewZoneProxy("sh004", "http://sh001-api.bilibili.com")
	group := e.Group("/x/player")
	{
		group.GET("/v2", authSvr.Guest, playerV2)
		group.GET("/policy", authSvr.Guest, policy)
		group.GET("/carousel.so", carousel)
		group.GET("/view", view)
		group.GET("/matsuri", matPage)
		group.GET("/pagelist", pageList)
		group.GET("/videoshot", authSvr.Guest, videoShot)
		group.GET("/playurl/token", authSvr.User, playURLToken)
		group.GET("/playurl", authSvr.Guest, playurl)
		group.POST("/card/click", proxyHandler, authSvr.User, playerCardClick)
		group.GET("/online/total", authSvr.Guest, onlineTotal)
		hls := group.Group("/hls")
		hls.GET("", authSvr.Guest, playurlHls)
		hls.GET("/master.m3u8", authSvr.Guest, hlsMaster)
		hls.GET("/stream.m3u8", authSvr.Guest, hlsStream)
	}
}

func internalRouter(e *bm.Engine) {
	group := e.Group("/x/internal/player")
	{
		group.GET("/playurl", vfySvr.Verify, authSvr.Guest, playurl)
		group.GET("/v2", authSvr.Guest, playerV2)
		hls := group.Group("/hls")
		hls.GET("", authSvr.Guest, playurlHls)
		hls.GET("/master.m3u8", authSvr.Guest, hlsMaster)
		hls.GET("/stream.m3u8", authSvr.Guest, hlsStream)
	}
}

func ping(c *bm.Context) {
	if err := playSvr.Ping(c); err != nil {
		log.Error("player service ping error(%v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func bvArgCheck(aid int64, bv string) (res int64, err error) {
	res = aid
	if bv != "" {
		if res, err = bvid.BvToAv(bv); err != nil {
			log.Error("View bvid.BvToAv(%s) aid(%d) error(%+v)", bv, aid, err)
			err = ecode.RequestErr
			return
		}
	}
	if res <= 0 {
		err = ecode.RequestErr
	}
	return
}
