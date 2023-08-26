package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/rate/limit/quota"

	"go-gateway/app/app-svr/archive/service/conf"
	"go-gateway/app/app-svr/archive/service/server/grpc"
	"go-gateway/app/app-svr/archive/service/server/http"
	"go-gateway/app/app-svr/archive/service/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()

	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("archive-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("archive-service.toml", cfg); err != nil {
		panic(err)
	}

	// init ecode
	ecode.Init(nil)
	// init log
	log.Init(cfg.Xlog)
	trace.Init(cfg.Tracer)
	defer trace.Close()
	defer log.Close()
	log.Info("archive-service start")
	quota.Init()
	defer quota.Close()
	// service init
	svr := service.New(cfg)
	// statsd init
	grpcSvr, err := grpc.New(nil, svr, cfg)
	if err != nil {
		panic(err)
	}
	http.Init(cfg, svr)

	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("archive-service get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.TODO())
			time.Sleep(time.Second * 2)
			log.Info("archive-service exit")
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
