package http

import (
	"go-common/library/conf/paladin"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/steins-gate/job/internal/service"
)

var (
	//nolint:unused
	svc *service.Service
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine) {
	var (
		hc struct {
			Server *bm.ServerConfig
		}
	)
	if err := paladin.Get("http.toml").UnmarshalTOML(&hc); err != nil {
		if err != paladin.ErrNotExist {
			panic(err)
		}
	}
	svc = s
	engine = bm.DefaultServer(nil)
	engine.Ping(ping)
	engine.Register(register)
	if err := engine.Start(); err != nil {
		panic(err)
	}
	return
}

func ping(ctx *bm.Context) {}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)

}
