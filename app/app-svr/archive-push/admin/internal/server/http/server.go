package http

import (
	"go-common/library/conf/paladin"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/http/blademaster/middleware/verify"

	"go-gateway/app/app-svr/archive-push/admin/internal/service"
)

var svc *service.Service
var p *permit.Permit
var vrf *verify.Verify

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
	p = permit.New2(nil)
	vrf = verify.New(nil)
	engine = bm.DefaultServer(&cfg)
	engine.Use(bm.CORS())
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/x/admin/archive-push")
	{
		// 推送批次
		g.GET("/batches", p.Permit2("SDK_ARCHIVE_PUSH_LIST"), batches)
		g.POST("/batches/push", p.Permit2("SDK_ARCHIVE_PUSH_ADMIN"), batchPush)
		g.GET("/batches/:id", p.Permit2("SDK_ARCHIVE_PUSH_ADMIN"), batchByID)
		g.GET("/batches/:id/export", p.Permit2("SDK_ARCHIVE_PUSH_LIST"), batchExportByID)
		g.GET("/batches/:id/export/initial", p.Permit2("SDK_ARCHIVE_PUSH_LIST"), initialBatchExportByID)

		// 稿件
		g.GET("/archives/pushed", p.Permit2("SDK_ARCHIVE_PUSH_LIST"), archivesPushed)
		g.POST("/archives/withdraw", p.Permit2("SDK_ARCHIVE_PUSH_ADMIN"), archiveWithdraw)

		// 作者
		g.GET("/authors", p.Permit2("SDK_ARCHIVE_PUSH_LIST"), authors)
		g.POST("/authors", p.Permit2("SDK_ARCHIVE_PUSH_ADMIN"), addAuthors)
		g.POST("/authors/remove", p.Permit2("SDK_ARCHIVE_PUSH_ADMIN"), removeAuthor)

		// 作者推送
		g.GET("/authors/pushBatches", p.Permit2("SDK_ARCHIVE_PUSH_LIST"), authorPushBatches)
		g.POST("/authors/pushBatches", p.Permit2("SDK_ARCHIVE_PUSH_ADMIN"), createAuthorPushBatches)
		g.POST("/authors/pushBatches/edit", p.Permit2("SDK_ARCHIVE_PUSH_ADMIN"), editAuthorPushBatches)
		g.POST("/authors/pushBatches/inactivate", p.Permit2("SDK_ARCHIVE_PUSH_ADMIN"), inactivateAuthorPushBatches)

		// 推送厂商
		g.GET("/vendors/available", p.Permit2("SDK_VENDOR_LIST"), vendorsAvailable)
		g.GET("/vendors/userBindable", p.Permit2("SDK_VENDOR_LIST"), vendorsUserBindable)

		// 对外（主要对网关）用接口
		apiG := g.Group("/api")
		{
			apiG.POST("/users/bindingSync", apiAuthorsBindingSync)
			apiG.GET("/users/binding", apiGetAuthors)
			apiG.POST("/archives/statusSync", apiArchivesStatusSync)
		}
	}
}

func ping(_ *bm.Context) {
}
