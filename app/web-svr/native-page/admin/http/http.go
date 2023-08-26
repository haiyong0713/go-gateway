package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"

	"go-gateway/app/web-svr/native-page/admin/conf"
	"go-gateway/app/web-svr/native-page/admin/service"
	"go-gateway/app/web-svr/native-page/admin/service/native"
)

var (
	actSrv    *service.Service
	authSrv   *permit.Permit
	natSrv    *native.Service
	verifySvc *verify.Verify
)

// Init init http sever instance.
func Init(c *conf.Config, s *service.Service) {
	actSrv = s
	verifySvc = verify.New(nil)
	natSrv = native.New(c)
	authSrv = permit.New2(nil)
	engine := bm.DefaultServer(c.HTTPServer)
	route(engine)
	if err := engine.Start(); err != nil {
		log.Error("httpx.Serve error(%v)", err)
		panic(err)
	}
}

func route(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/admin/native_page")
	{
		napp := g.Group("/native")
		{
			napp.POST("/page/add", addPage)
			napp.POST("/page/modify", modifyPage)
			napp.POST("/page/online", reOnline)
			napp.POST("/page/del", delPage)
			napp.POST("/page/edit", editPage)
			napp.GET("/page/search", searchPage)
			napp.GET("/up/page", upPage)
			napp.GET("/page/find", findPage)
			napp.GET("/page/module/search", searchModule)
			napp.POST("/page/module/save", saveModule)
			napp.POST("/tab/save", authSrv.Permit2(""), saveTab)
			napp.POST("/tab/edit", authSrv.Permit2(""), editTab)
			napp.GET("/tab/list", authSrv.Permit2(""), tabList)
			napp.GET("/tab/page_tab", pageTab)
			napp.POST("/ts/online", verifySvc.Verify, tsOnline)
			napp.GET("/counters", findCounters)
			napp.POST("/ts/space/offline", verifySvc.Verify, spaceOffline)
			napp.POST("/topic/upgrade", verifySvc.Verify, topicUpgrade)
			napp.GET("/game/detail", gameDetail)
			napp.GET("/cartoon/detail", cartoonDetail)
			napp.GET("/channel/detail", channelDetail)
			napp.GET("/reserve/detail", reserveDetail)
			napp.GET("/ts/page", tsPage)
			napp.GET("/up_vote", upVote)
			napp.POST("/newact/add", verifySvc.Verify, addNewact)
		}
		groupWhite := g.Group("/white_list")
		{
			groupWhite.POST("/add", authSrv.Permit2(""), addWhiteList)
			groupWhite.POST("/add/batch", authSrv.Permit2(""), batchAddWhiteList)
			groupWhite.POST("/add/outer", verifySvc.Verify, addWhiteListOuter)
			groupWhite.POST("/delete", authSrv.Permit2(""), deleteWhiteList)
			groupWhite.GET("/list", authSrv.Permit2(""), whiteList)
		}
	}
}

func ping(c *bm.Context) {
	if err := actSrv.Ping(c); err != nil {
		c.Error = err
		c.AbortWithStatus(503)
	}
}
