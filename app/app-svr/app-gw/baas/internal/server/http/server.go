package http

import (
	"net/http"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	pb "go-gateway/app/app-svr/app-gw/baas/api"
	"go-gateway/app/app-svr/app-gw/baas/internal/service"
)

var (
	svc       *service.Service
	permitSvc *permit.Permit
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
	engine = bm.DefaultServer(&cfg)
	initService()
	pb.RegisterBaasBMServer(engine, s)
	initRouter(engine)
	err = engine.Start()
	return
}
func initService() {
	permitSvc = permit.New2(nil)
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	root := e.Group("/x/baas")
	manager := root.Group("/admin", permitSvc.Verify2())
	{
		manager.GET("/authz", authz)

		mapper := manager.Group("/mapper")
		mapper.GET("/list", mapperModelList)
		mapper.GET("/item/list", mapperItemList)
		mapper.GET("/detail", mapperModelDetail)
		mapper.POST("/add", svc.Role.RoleAuthZ(), addMapperModel)
		mapper.GET("/field/list", mapperModelFieldList)
		mapper.POST("/field/add", svc.Role.RoleAuthZ(), addMapperModelField)
		mapper.POST("/field/update", svc.Role.RoleAuthZ(), updateMapperModelField)
		mapper.POST("/field/delete", svc.Role.RoleAuthZ(), deleteMapperModelField)
		mapper.POST("/field/rule/add", svc.Role.RoleAuthZ(), addMapperModelFieldRule)
		mapper.POST("/field/rule/update", svc.Role.RoleAuthZ(), updateMapperModelFieldRule)
	}
	manager.GET("/export/list", exportList)
	manager.POST("/export/add", svc.Role.RoleAuthZ(), addExport)
	manager.POST("/export/update", svc.Role.RoleAuthZ(), updateExport)

	manager.POST("/import/add", svc.Role.RoleAuthZ(), addImport)
	manager.POST("/import/update", svc.Role.RoleAuthZ(), updateImport)

	root.GET("/impl/*path", baasImpl)
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
