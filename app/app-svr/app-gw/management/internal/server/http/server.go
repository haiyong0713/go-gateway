package http

import (
	"go-common/library/conf/paladin.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/service"
)

var rawSvc *service.Service

// New new a bm server.
func New(s pb.ManagementServer) (engine *bm.Engine, err error) {
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
	rawSvc = s.(*service.Service)
	engine = bm.DefaultServer(&cfg)
	engine.Use(bm.CORS())
	pb.RegisterManagementBMServer(engine, s)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	grpc := &grpcServer{}
	http := &httpServer{}
	permit := permit.New2(nil)
	verify := verify.New(nil)
	e.Ping(ping)
	g := e.Group("/x/admin/app-gw", permit.Verify2())
	{
		g.GET("/authz", authz)
		g.GET("/authz-sidebar", authzSidebar)

		g.GET("/gateway", gateway)
		g.GET("/gateway/profile", gatewayProfile)
		g.POST("/gateway/add", rawSvc.Common.RoleAuthZ(), addGateway)
		g.POST("/gateway/update", rawSvc.Common.RoleAuthZ(), updateGateway)
		g.POST("/gateway/delete", rawSvc.Common.RoleAuthZ(), deleteGateway)
		g.POST("/gateway/config/init", initGatewayConfigs)
		g.POST("/gateway/config/enable/all", rawSvc.Common.RoleAuthZ(), enableAllGateway)
		g.POST("/gateway/config/disable/all", rawSvc.Common.RoleAuthZ(), disableAllGateway)
		g.POST("/gateway/grpc/config/enable/all", rawSvc.Common.RoleAuthZ(), enableAllGRPCGateway)
		g.POST("/gateway/grpc/config/disable/all", rawSvc.Common.RoleAuthZ(), disableAllGRPCGateway)
		g.GET("/gateway/proxy/:token/metrics", gatewayProxy)
		g.GET("/gateway/proxy/:token/configs.toml", gatewayProxy)
		g.GET("/gateway/proxy/:token/metrics.json", gatewayProxy)
		g.GET("/gateway/proxy/:token/grpc-configs.toml", gatewayProxy)

		g.GET("/dynpath", http.listDynPath)
		g.POST("/dynpath/add", rawSvc.Common.RoleAuthZ(), http.addDynPath)
		g.POST("/dynpath/delete", rawSvc.Common.RoleAuthZ(), http.deleteDynPath)
		g.POST("/dynpath/update", rawSvc.Common.RoleAuthZ(), http.updateDynPath)
		g.POST("/dynpath/enable", rawSvc.Common.RoleAuthZ(), http.enableDynPath)
		g.POST("/dynpath/disable", rawSvc.Common.RoleAuthZ(), http.disableDynPath)

		g.GET("/http/dynpath", http.listDynPath)
		g.POST("/http/dynpath/add", rawSvc.Common.RoleAuthZ(), http.addDynPath)
		g.POST("/http/dynpath/delete", rawSvc.Common.RoleAuthZ(), http.deleteDynPath)
		g.POST("/http/dynpath/update", rawSvc.Common.RoleAuthZ(), http.updateDynPath)
		g.POST("/http/dynpath/enable", rawSvc.Common.RoleAuthZ(), http.enableDynPath)
		g.POST("/http/dynpath/disable", rawSvc.Common.RoleAuthZ(), http.disableDynPath)

		g.GET("/grpc/dynpath", grpc.listDynPath)
		g.POST("/grpc/dynpath/add", rawSvc.Common.RoleAuthZ(), grpc.addDynPath)
		g.POST("/grpc/dynpath/delete", rawSvc.Common.RoleAuthZ(), grpc.deleteDynPath)
		g.POST("/grpc/dynpath/update", rawSvc.Common.RoleAuthZ(), grpc.updateDynPath)
		g.POST("/grpc/dynpath/enable", rawSvc.Common.RoleAuthZ(), grpc.enableDynPath)
		g.POST("/grpc/dynpath/disable", rawSvc.Common.RoleAuthZ(), grpc.disableDynPath)

		g.GET("/breaker/api", http.listBreakerAPI)
		g.POST("/breaker/add", rawSvc.Common.RoleAuthZ(), http.addBreakerAPI)
		g.POST("/breaker/update", rawSvc.Common.RoleAuthZ(), http.updateBreakerAPI)
		g.POST("/breaker/enable", rawSvc.Common.RoleAuthZ(), http.enableBreakerAPI)
		g.POST("/breaker/disable", rawSvc.Common.RoleAuthZ(), http.disableBreakerAPI)
		g.POST("/breaker/delete", rawSvc.Common.RoleAuthZ(), http.deleteBreakerAPI)

		g.GET("/http/breaker/api", http.listBreakerAPI)
		g.POST("/http/breaker/add", rawSvc.Common.RoleAuthZ(), http.addBreakerAPI)
		g.POST("/http/breaker/update", rawSvc.Common.RoleAuthZ(), http.updateBreakerAPI)
		g.POST("/http/breaker/enable", rawSvc.Common.RoleAuthZ(), http.enableBreakerAPI)
		g.POST("/http/breaker/disable", rawSvc.Common.RoleAuthZ(), http.disableBreakerAPI)
		g.POST("/http/breaker/delete", rawSvc.Common.RoleAuthZ(), http.deleteBreakerAPI)

		g.GET("/grpc/breaker/api", grpc.listBreakerAPI)
		g.POST("/grpc/breaker/add", rawSvc.Common.RoleAuthZ(), grpc.addBreakerAPI)
		g.POST("/grpc/breaker/update", rawSvc.Common.RoleAuthZ(), grpc.updateBreakerAPI)
		g.POST("/grpc/breaker/enable", rawSvc.Common.RoleAuthZ(), grpc.enableBreakerAPI)
		g.POST("/grpc/breaker/disable", rawSvc.Common.RoleAuthZ(), grpc.disableBreakerAPI)
		g.POST("/grpc/breaker/delete", rawSvc.Common.RoleAuthZ(), grpc.deleteBreakerAPI)

		g.GET("/snapshot/profile", snapshotProfile)
		g.POST("/snapshot/add", addSnapshot)
		g.POST("/snapshot/impl/:snapshot_id/:resource/:action", snapshotAction)
		g.POST("/snapshot/impl-http/:snapshot_id/:resource/:action", snapshotAction)
		g.POST("/snapshot/impl-grpc/:snapshot_id/:resource/:action", snapshotGRPCAction)

		g.GET("/deployment", deployment)
		g.POST("/deployment/create", createDeployment)
		g.GET("/deployment/compare", compareDeployment)
		g.POST("/deployment/confirm", confirmDeployment)
		g.GET("/deployment/deploy/profile", deployDeploymentProfile)
		g.POST("/deployment/deploy", deployDeployment)
		g.POST("/deployment/all/deploy", deployDeploymentAll)
		g.POST("/deployment/finish", finishDeployment)
		g.POST("/deployment/rollback", rollbackDeployment)
		g.POST("/deployment/close", closeDeployment)
		g.POST("/deployment/cancel", cancelDeployment)

		g.GET("/prompt/app", appPromptAPI)
		g.GET("/prompt/gateway/config", configPromptAPI)
		g.GET("/prompt/app/path", appPathPromptAPI)
		g.GET("/prompt/zone", zonePromptAPI)
		g.GET("/prompt/grpc/app/method", grpcAppMethodPromptAPI)
		g.GET("/prompt/grpc/app/package", grpcAppPackagePromptAPI)

		g.GET("/limiter", listLimiter)
		g.POST("/limiter/add", addLimiter)
		g.POST("/limiter/update", updateLimiter)
		g.POST("/limiter/delete", deleteLimiter)
		g.POST("/limiter/enable", enableLimiter)
		g.POST("/limiter/disable", disableLimiter)

		g.POST("/plugin/setup", setupPlugin)
		g.GET("/plugin/list", pluginList)
		g.GET("/log", listLog)
		g.POST("/task/execute", executeTask)
	}
	p := e.Group("/x/admin/app-gw/internal", rawSvc.Common.AuthByAppKey(), verify.Verify)
	{
		p.GET("/http/dynpath", http.listDynPath)
		p.POST("/http/dynpath/add", http.addDynPath)
		p.POST("/http/dynpath/delete", http.deleteDynPath)
		p.POST("/http/dynpath/update", http.updateDynPath)
		p.POST("/http/dynpath/enable", http.enableDynPath)
		p.POST("/http/dynpath/disable", http.disableDynPath)

		p.GET("/http/breaker/api", http.listBreakerAPI)
		p.POST("/http/breaker/add", http.addBreakerAPI)
		p.POST("/http/breaker/update", http.updateBreakerAPI)
		p.POST("/http/breaker/enable", http.enableBreakerAPI)
		p.POST("/http/breaker/disable", http.disableBreakerAPI)
		p.POST("/http/breaker/delete", http.deleteBreakerAPI)

		p.GET("/grpc/dynpath", grpc.listDynPath)
		p.POST("/grpc/dynpath/add", grpc.addDynPath)
		p.POST("/grpc/dynpath/delete", grpc.deleteDynPath)
		p.POST("/grpc/dynpath/update", grpc.updateDynPath)
		p.POST("/grpc/dynpath/enable", grpc.enableDynPath)
		p.POST("/grpc/dynpath/disable", grpc.disableDynPath)

		p.GET("/grpc/breaker/api", grpc.listBreakerAPI)
		p.POST("/grpc/breaker/add", grpc.addBreakerAPI)
		p.POST("/grpc/breaker/update", grpc.updateBreakerAPI)
		p.POST("/grpc/breaker/enable", grpc.enableBreakerAPI)
		p.POST("/grpc/breaker/disable", grpc.disableBreakerAPI)
		p.POST("/grpc/breaker/delete", grpc.deleteBreakerAPI)
	}

}

func ping(ctx *bm.Context) {}
