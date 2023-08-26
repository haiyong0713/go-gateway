package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/app-svr/resource/service/conf"
	"go-gateway/app/app-svr/resource/service/http"
	rpc "go-gateway/app/app-svr/resource/service/rpc/server"
	grpc "go-gateway/app/app-svr/resource/service/server/grpc"
	"go-gateway/app/app-svr/resource/service/service"
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
	if err = paladin.Get("resource-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("resource-service.toml", cfg); err != nil {
		panic(err)
	}
	// init log
	log.Init(cfg.XLog)
	trace.Init(cfg.Tracer)
	defer trace.Close()
	defer log.Close()
	log.Info("resource-service start")
	// service init
	svr := service.New(cfg)
	rpcSvr := rpc.New(cfg, svr)
	grpcSvr := grpc.New(nil, svr)
	http.Init(cfg, svr)
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("resource-service get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			rpcSvr.Close()
			if err := grpcSvr.Shutdown(context.Background()); err != nil {
				log.Error("resource-service grpcSvr shutdown error: %s", err.Error())
			}
			svr.Close()
			time.Sleep(time.Second * 2)
			log.Info("resource-service exit")
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
