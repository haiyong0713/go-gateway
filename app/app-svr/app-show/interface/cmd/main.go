package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/trace"
	_ "go.uber.org/automaxprocs"

	"go-gateway/app/app-svr/app-show/interface/component"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/http"
	"go-gateway/app/app-svr/app-show/interface/server/grpc"
	"go-gateway/app/app-svr/app-show/interface/service/act"
	"go-gateway/app/app-svr/app-show/interface/service/banner"
	"go-gateway/app/app-svr/app-show/interface/service/daily"
	"go-gateway/app/app-svr/app-show/interface/service/ping"
	"go-gateway/app/app-svr/app-show/interface/service/rank"
	"go-gateway/app/app-svr/app-show/interface/service/rank-list"
	"go-gateway/app/app-svr/app-show/interface/service/region"
	"go-gateway/app/app-svr/app-show/interface/service/show"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

func main() {
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("app-show.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("app-show.toml", cfg); err != nil {
		panic(err)
	}
	// init ecode
	ecode.Init(nil)
	// init log
	log.Init(cfg.XLog)
	defer log.Close()
	log.Info("app-show start")
	// init trace
	trace.Init(cfg.Tracer)
	defer trace.Close()
	if err := component.InitByCfg(cfg); err != nil {
		log.Error("sql.InitByCfg error(%v)", err)
		panic(err)
	}
	svr := initService(cfg)
	grpc.New(cfg.RpcServer, svr)
	// service init
	http.Init(cfg, svr)
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-show get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			http.Close()
			component.Close()
			log.Info("app-show exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}

// initService init services.
func initService(c *conf.Config) (svr *http.Server) {
	svr = &http.Server{
		AuthSvc:     auth.New(nil),
		BannerSvc:   banner.New(c),
		RegionSvc:   region.New(c),
		ShowSvc:     show.New(c),
		PingSvc:     ping.New(c),
		RankSvc:     rank.New(c),
		DailySvc:    daily.New(c),
		ActSvr:      act.New(c),
		RankListSvc: ranklist.New(c),
		VerifySvr:   verify.New(nil),
		Config:      c,
		FeatureSvr:  feature.New(nil),
	}
	return
}
