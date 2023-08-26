package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/aurora"
	"go-common/library/conf/env"
	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-player/interface/conf"
	"go-gateway/app/app-svr/app-player/interface/grpc"
	"go-gateway/app/app-svr/app-player/interface/http"
	"go-gateway/app/app-svr/app-player/interface/service"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("app-player.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("app-player.toml", cfg); err != nil {
		panic(err)
	}
	// init log
	log.Init(cfg.Log)
	defer log.Close()
	log.Info("app-player start")
	// init trace
	if env.DeployEnv == env.DeployEnvProd {
		trace.Init(nil)
		defer trace.Close()
	}
	// ecode init
	ecode.Init(cfg.Ecode)
	anticrawler.Init(nil)
	aurora.Init(nil)
	// service init
	featureSvr := feature.New(nil)
	http.Init(cfg, featureSvr)
	svc := service.New(cfg)
	// grpc init
	grpcSvr, err := grpc.New(cfg.RpcServer, svc, featureSvr)
	if err != nil {
		panic(err)
	}
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-player get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.TODO())
			svc.Close()
			log.Info("app-player exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
