package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/web-svr/appstatic/admin/conf"
	"go-gateway/app/web-svr/appstatic/admin/service"
	"go-gateway/app/web-svr/appstatic/admin/service/peak"
)

var (
	vfySvc  *verify.Verify
	authSvc *permit.Permit
	apsSvc  *service.Service
	peakSvc *peak.Service
)

// Init http server
func Init(c *conf.Config, s *service.Service) {
	initService(c, s)
	engine := bm.DefaultServer(c.BM)
	innerRouter(engine)
	// init internal server
	if err := engine.Start(); err != nil {
		log.Error("engine.Start error(%v)", err)
		panic(err)
	}
}

// initService init service
func initService(c *conf.Config, s *service.Service) {
	apsSvc = s
	authSvc = permit.New2(nil)
	vfySvc = verify.New(nil)
	peakSvc = peak.New(c)
}

// innerRouter
func innerRouter(e *bm.Engine) {
	// ping monitor
	e.GET("/monitor/ping", ping)
	e.Use(bm.CORS())
	// internal api
	bg := e.Group("/x/admin/appstatic/res")
	{
		bg.POST("/add_ver", authSvc.Permit2("APP_RESOURCE_POOL_MGT"), addVer)           // 从mgr上传，正式权限
		bg.POST("/add_ver_test", authSvc.Permit2("APP_RESOURCE_POOL_MGT_EDIT"), addVer) // 从mgr上传，测试权限
		bg.POST("/upload", vfySvc.Verify, addVer)                                       // 从其他系统上传
		bg.POST("/publish", vfySvc.Verify, publish)                                     // 告知某资源包的第一次发布，用于触发增量包补充计算
		bg.POST("/push", authSvc.Permit2("APP_RESOURCE_POOL_MGT"), push)                //立即推送
		bg.GET("/boss/publish/check", cdnPublishCheck)                                  //资源发布检测
		bg.GET("/boss/url/status", cdnStatus)                                           //资源预热状态查询
		bg.POST("/boss/url/preload", cdnPreload)                                        //预热大文件
	}
	gray := e.Group("/x/admin/appstatic/gray", authSvc.Permit2("APP_RESOURCE_GRAY"))
	{
		gray.GET("", grayIndex)
		gray.POST("/add", addGray)
		gray.POST("/update", saveGray)
		gray.POST("/upload", uploadGray)
	}
	chronos := e.Group("x/admin/appstatic/chronos", authSvc.Permit2("APP_RESOURCE_CHRONOS"))
	{
		chronos.POST("/upload", uploadChronos)
		chronos.POST("/save", saveChronos)
		chronos.GET("/list", listChronos)
	}
	chronosV2 := e.Group("x/admin/appstatic/chronos/v2")
	{
		chronosV2App := chronosV2.Group("/app", chronosRsaVerify)
		{
			chronosV2App.POST("/save", saveChronosV2App)
			chronosV2App.POST("/delete", deleteChronosV2App)
			chronosV2App.GET("/list", showChronosV2AppList)
			chronosV2App.GET("/detail", showChronosV2AppDetail)
		}
		chronosV2Service := chronosV2.Group("/service", chronosRsaVerify)
		{
			chronosV2Service.POST("/save", saveChronosV2Service)
			chronosV2Service.POST("/delete", deleteChronosV2Service)
			chronosV2Service.GET("/list", showChronosV2ServiceList)
			chronosV2Service.GET("/detail", showChronosV2ServiceDetail)
		}
		chronosV2Package := chronosV2.Group("/package", chronosRsaVerify)
		{
			chronosV2Package.POST("/save", saveChronosV2Package, authSvc.Permit2("AUDIT_RESOURCE_CHRONOS"))
			chronosV2Package.POST("/rank", rankChronosV2Package, authSvc.Permit2("AUDIT_RESOURCE_CHRONOS"))
			chronosV2Package.POST("/delete", deleteChronosV2Package, authSvc.Permit2("AUDIT_RESOURCE_CHRONOS"))
			chronosV2Package.GET("/list", showChronosV2Package)
			chronosV2Package.GET("/detail", showChronosV2PackageDetail)
			chronosV2Package.POST("/upload", uploadChronos)
			chronosV2Package.POST("/batch/save", batchSaveChronosV2Packages)
		}
		chronosV2Audit := chronosV2.Group("/audit")
		{
			chronosV2Audit.POST("/approved", approved)
			chronosV2Audit.POST("/reject", reject)
			chronosV2Audit.GET("/list", auditList)
		}
	}
	videoResolutionCtrl := e.Group("x/admin/appstatic/resolution")
	{
		dolby := videoResolutionCtrl.Group("/dolby/whitelist")
		{
			dolby.GET("", dolbyWhiteList)
			dolby.POST("/add", addDolbyWhiteList)
			dolby.POST("/save", saveDolbyWhiteList)
			dolby.POST("/delete", deleteDolbyWhiteList)
		}
		qn := videoResolutionCtrl.Group("/qn/blacklist")
		{
			qn.GET("", qnBlackList)
			qn.POST("/add", addQnBlackList)
			qn.POST("/save", saveQnBlackList)
			qn.POST("/delete", deleteQnBlackList)
		}
		limitFree := videoResolutionCtrl.Group("/limit/free")
		{
			limitFree.GET("/list", limitFreeList)
			limitFree.POST("/add", addLimitFree)
			limitFree.POST("/edit", editLimitFree)
			limitFree.POST("/delete", deleteLimitFree)
		}

	}
	peak := e.Group("/x/admin/appstatic/peak")
	{
		peak.POST("/add", addPeak)
		peak.GET("/list", indexPeak)
		peak.POST("/update", updatePeak)
		peak.POST("/publish", publishPeak)
		peak.POST("/delete", deletePeak)
		peak.POST("/upload", uploadPeak)
	}
	e.POST("/x/admin/appstatic/upload/big", authSvc.Permit2("APP_RESOURCE_BIGFILE"), addBigVer)
}

// ping check server ok.
func ping(c *bm.Context) {}
