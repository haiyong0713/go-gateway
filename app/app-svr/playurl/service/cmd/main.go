package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/aurora"
	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"

	"go-gateway/app/app-svr/playurl/service/conf"
	"go-gateway/app/app-svr/playurl/service/server/grpc"
	"go-gateway/app/app-svr/playurl/service/server/http"
	"go-gateway/app/app-svr/playurl/service/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()

	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("playurl-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("playurl-service.toml", cfg); err != nil {
		panic(err)
	}

	wardenSDKCfg := &conf.WardenSDKConfig{}
	if err := paladin.Get("grpc-client-sdk.toml").UnmarshalTOML(&wardenSDKCfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("grpc-client-sdk.toml", wardenSDKCfg); err != nil {
		panic(err)
	}

	log.Init(cfg.Log)
	defer log.Close()
	log.Info("playurl-service start")
	aurora.Init(nil)
	trace.Init(cfg.Tracer)
	defer trace.Close()
	ecode.Init(cfg.Ecode)
	svc := service.New(cfg)
	grpcServ := grpc.New(cfg.GRPC, svc)
	http.Init(cfg, svc)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			_ = grpcServ.Shutdown(ctx)
			log.Info("playurl-service exit")
			svc.Close()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
