package http

import (
	"go-common/library/conf/paladin"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	pb "go-gateway/app/web-svr/datasource-ng/admin/api"
)

var (
	svc       pb.DatasourceServer
	permitSvc *permit.Permit
)

// New new a bm server.
func New(s pb.DatasourceServer) (engine *bm.Engine, err error) {
	var (
		hc struct {
			Server *bm.ServerConfig
		}
	)
	if err = paladin.Get("http.toml").UnmarshalTOML(&hc); err != nil {
		if err != paladin.ErrNotExist {
			return
		}
		err = nil
	}
	svc = s
	engine = bm.DefaultServer(hc.Server)
	initService()
	initRouter(engine)
	err = engine.Start()
	return
}

func initService() {
	permitSvc = permit.New2(nil)
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	root := e.Group("/x/admin/datasource")
	model := root.Group("/model")
	{
		model.GET("list", ModelList)
		model.GET("all", ModelAll)
		model.GET("detail", ModelDetail)
		model.POST("create", permitSvc.Verify2(), ModelCreate)
	}
	item := model.Group("/item")
	{
		item.GET("list", ModelItemList)
		item.GET("detail", ModelItemDetail)
		item.POST("create", permitSvc.Verify2(), ModelItemCreate)
	}
}

func ping(ctx *bm.Context) {}
