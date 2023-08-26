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
	"go-gateway/app/app-svr/archive-honor/service/conf"
	"go-gateway/app/app-svr/archive-honor/service/server/grpc"
	"go-gateway/app/app-svr/archive-honor/service/server/http"
	"go-gateway/app/app-svr/archive-honor/service/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("archive-honor-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("archive-honor-service.toml", cfg); err != nil {
		panic(err)
	}
	// init ecode
	ecode.Init(nil)
	// init log
	log.Init(cfg.Log)
	trace.Init(nil)
	defer trace.Close()
	defer log.Close()
	log.Info("archive-honor-service start")
	svr := service.New(cfg)
	grpcSvr, err := grpc.New(nil, svr, cfg)
	if err != nil {
		panic(err)
	}
	http.Init(cfg)
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("archive-honor-service get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.TODO())
			time.Sleep(time.Second * 2)
			svr.Close()
			log.Info("archive-honor-service exit")
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
