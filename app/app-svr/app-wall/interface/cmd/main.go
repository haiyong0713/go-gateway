package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/trace"
	"go-common/library/queue/databus/actionlog"
	"go-common/library/queue/databus/report"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	"go-gateway/app/app-svr/app-wall/interface/grpc"
	"go-gateway/app/app-svr/app-wall/interface/http"
	"go-gateway/app/app-svr/app-wall/interface/service/mobile"
	"go-gateway/app/app-svr/app-wall/interface/service/offer"
	"go-gateway/app/app-svr/app-wall/interface/service/operator"
	pingsvr "go-gateway/app/app-svr/app-wall/interface/service/ping"
	"go-gateway/app/app-svr/app-wall/interface/service/telecom"
	"go-gateway/app/app-svr/app-wall/interface/service/unicom"
	"go-gateway/app/app-svr/app-wall/interface/service/wall"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("app-wall.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("app-wall.toml", cfg); err != nil {
		panic(err)
	}
	// init log
	log.Init(cfg.Log)
	defer log.Close()
	log.Info("app-wall start")
	// init trace
	trace.Init(cfg.Tracer)
	defer trace.Close()
	anticrawler.Init(nil)
	// ecode init
	ecode.Init(cfg.Ecode)
	// report init
	report.InitUser(cfg.Report)
	// actionlog init
	actionlog.InitUser(nil)
	// service init
	svr := initService(cfg)
	http.Init(cfg, svr)
	grpcSvr, err := grpc.New(nil, svr)
	if err != nil {
		panic(err)
	}
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-wall get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if err := grpcSvr.Shutdown(context.TODO()); err != nil {
				log.Error("%+v", err)
			}
			http.Close()
			log.Info("app-wall exit")
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
		VerifySvc: verify.New(nil),
		AuthSvc:   auth.New(&auth.Config{DisableCSRF: true}),
		// init self service
		WallSvc:     wall.New(c),
		OfferSvc:    offer.New(c),
		UnicomSvc:   unicom.New(c),
		MobileSvc:   mobile.New(c),
		PingSvc:     pingsvr.New(c),
		TelecomSvc:  telecom.New(c),
		OperatorSvc: operator.New(c),
	}
	return
}
