package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/macross/service/conf"
	"go-gateway/app/app-svr/macross/service/service"
)

var (
	verifySvc *verify.Verify
	permitSvc *permit.Permit
	svr       *service.Service
)

// Init int http service
func Init(c *conf.Config) {
	verifySvc = verify.New(nil)
	permitSvc = permit.New2(nil)
	svr = service.New(conf.Conf)
	// init internal router
	engineInner := bm.DefaultServer(c.BM.Inner)
	innerRouter(engineInner)
	// init internal server
	if err := engineInner.Start(); err != nil {
		log.Error("engineInner.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
	// init external router
	engineLocal := bm.DefaultServer(c.BM.Local)
	localRouter(engineLocal)
	// init external server
	if err := engineLocal.Start(); err != nil {
		log.Error("engineLocal.Start() error(%v) | config(%v)", err, c)
		panic(err)
	}
}

// innerRouter init outer router api path.
func innerRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Register(register)
	rs := e.Group("/api/v2/macross")
	// MANAGER
	mng := rs.Group("/manager")
	// auth init.
	mng.GET("/getAuths", permitSvc.Verify2(), getAuths)
	// user.
	mng.GET("/user", user)
	mng.POST("/user/save", saveUser)
	mng.POST("/user/del", delUser)
	// role.
	mng.GET("/role", role)
	mng.POST("/role/save", saveRole)
	mng.POST("/role/del", DelRole)
	// auth.
	mng.GET("/auth", auth)
	mng.POST("/auth/save", saveAuth)
	mng.POST("/auth/del", delAuth)
	// relation
	mng.POST("/setRelation", authRelation)
	// dashboard
	rs.POST("/dashboard", dashboard)
	// sendmail
	rs.POST("/sendmail", sendmail)
	// package upload
	rs.POST("/upload", packageUpload)
	// get package list
	rs.GET("/archive", packageList)
	// android
	android := rs.Group("/android")
	android.POST("/upload", apklUpload)
	android.POST("/upload/cdn", apklUploadCDN)
}

// localRouter init local router api path.
func localRouter(e *bm.Engine) {
	e.GET("/x/macross/version", version)
}
