package http

import (
	"go-common/library/conf/paladin"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/web-svr/datasource-ng/service/api"
)

var (
	svc pb.DataSourceNGServer
)

// New new a bm server.
func New(s pb.DataSourceNGServer) (engine *bm.Engine, err error) {
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
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	ig := e.Group("/x/internal/datasource-ng")
	og := e.Group("/x/datasource-ng")
	setupCommonAPI(ig)
	setupCommonAPI(og)
}

func setupCommonAPI(g bm.IRoutes) {
	g.GET("/flat/item", flatItem)
	g.GET("/item", item)
}

func ping(ctx *bm.Context) {}
