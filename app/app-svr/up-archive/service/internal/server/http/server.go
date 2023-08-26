package http

import (
	"net/http"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-gateway/app/app-svr/up-archive/service/api"
	"go-gateway/app/app-svr/up-archive/service/internal/service"

	"github.com/pkg/errors"
)

var (
	svc    *service.Service
	idfSvc *verify.Verify
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
	var (
		ct       paladin.TOML
		cfg      bm.ServerConfig
		quotaCfg *quota.Config
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = ct.Get("quota").UnmarshalTOML(&quotaCfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	svc = s
	idfSvc = verify.New(nil)
	engine = bm.DefaultServer(&cfg)
	limiter := quota.New(quotaCfg)
	engine.Use(limiter.Handler())
	api.RegisterUpArchiveBMServer(engine, s)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	e.Inject(api.PathUpArchiveArcPassed, idfSvc.Verify)
	e.Inject(api.PathUpArchiveArcPassedTotal, idfSvc.Verify)
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
