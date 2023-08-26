package http

import (
	"go-common/library/conf/paladin"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-gateway/app/api-gateway/api-manager/internal/service"
)

var (
	svc     *service.Service
	authSvc *permit.Permit
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
	authSvc = permit.New2(nil)
	engine = bm.DefaultServer(&cfg)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/admin/api-gateway")
	{
		g.GET("/discovery_list", discoveryList)
		cg := g.Group("/code")
		{
			cg.POST("/generate", codeGenerate)
		}
		ag := g.Group("/api")
		{
			ag.GET("/list", appList)
			ag.POST("/add", addApi)
			ag.POST("/edit", editApi)
			ag.POST("/delete", delApi)
		}
		wfg := g.Group("/workflow")
		{
			wfg.GET("/status", wfStatus)
			wfg.POST("/create", createWF)
			wfg.POST("/resume", resumeWF)
			wfg.POST("/manual_stop", stopWF)
		}
		contral := g.Group("/contral", authSvc.Verify2())
		{
			ctGroup := contral.Group("/group")
			{
				ctGroup.POST("/add", groupAdd)                    // 添加分组
				ctGroup.POST("/edit", groupEdit)                  // 编辑分组
				ctGroup.GET("/list", groupList)                   // 分组列表
				ctGroup.POST("/follow/action", groupFollowAction) // 分组关注/取消关注操作
				ctGroup.GET("/follow/list", groupFollowList)      // 我关注的分组列表
			}
			ctApi := contral.Group("/api")
			{
				ctApi.POST("/add", apiAdd)
				ctApi.POST("/edit", apiEdit)
				ctApi.GET("/list", apiList)
				ctApi.POST("/config/add", apiConfigAdd)
				ctApi.POST("/config/rollback", apiConfigRollback)
				ctApi.GET("/config/list", apiConfigList)
				ctApi.GET("/publish/list", apiPublishList)
				ctApi.POST("/publish/callback", apiPublishCallback)
			}
		}
	}
}

func ping(ctx *bm.Context) {
}
