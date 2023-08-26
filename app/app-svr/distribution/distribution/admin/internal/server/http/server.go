package http

import (
	"net/http"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/service"
	"go-gateway/app/app-svr/distribution/distribution/admin/logcontext"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
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
	e.Use(bm.CORS())
	distribution := e.Group("/x/admin/distribution")
	distribution.POST("/rename", Rename)
	distribution.GET("/log/action", actionLog)
	{
		abtest := distribution.Group("/abtest")
		{
			abtest.GET("/list", ABTestList)
			abtest.GET("/detail", ABTestDetail)
			abtest.POST("/save", logcontext.ActionLogHandler(), ABTestSave)
		}
		tusSingle := distribution.Group("/tus")
		{
			//查看所有人群包信息的接口
			tusSingle.GET("/list", TusList)
			tusSingle.GET("/detail", TusDetail)
			//保存详细内容
			tusSingle.POST("/save", TusSave)
		}
		tusMultiple := distribution.Group("/tus/multiple")
		{
			tusMultiple.GET("/fields", MultipleTusFields)
			tusMultiple.GET("/detail", MultipleTusDetail)
			tusMultiple.POST("/save", logcontext.ActionLogHandler(), MultipleTusSave)
		}
		tusEdit := distribution.Group("/tus/multiple/edit")
		{
			tusEdit.GET("/overview", Overview)
			tusEdit.GET("/performance", Performance)
			tusEdit.POST("/performance/save", logcontext.ActionLogHandler(), PerformanceSave)
		}
		tusVersion := distribution.Group("/tus/multiple/version")
		{
			tusVersion.POST("/add", logcontext.ActionLogHandler(), tusMultipleVersionAdd)
			tusVersion.POST("/update/buildlmit", logcontext.ActionLogHandler(), tusMultipleUpdateBuildLimit)
			tusVersion.GET("/list", tusMultipleVersion)
			tusVersion.POST("/delete", logcontext.ActionLogHandler(), tusMultipleVersionDelete)
		}
	}
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
