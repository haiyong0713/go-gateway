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
	"go-common/library/net/trace"
	"go-common/library/queue/databus/report"
	"go-common/library/rate/limit/quota"

	"go-gateway/app/web-svr/native-page/interface/conf"
	"go-gateway/app/web-svr/native-page/interface/http"
	"go-gateway/app/web-svr/native-page/interface/server/grpc"
	"go-gateway/app/web-svr/native-page/interface/service/like"

	"go-main/app/archive/aegis/admin/server/databus"

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
	if err = paladin.Get("native-page-interface.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("native-page-interface.toml", cfg); err != nil {
		panic(err)
	}
	// init log
	log.Init(cfg.Log)
	trace.Init(cfg.Tracer)
	quota.Init()
	defer func() {
		quota.Close()
		_ = log.Close()
		_ = trace.Close()
	}()
	log.Info("native-page-interface start")
	// ecode
	ecode.Init(cfg.Ecode)
	// report
	report.InitUser(nil)
	// 送审通知sdk
	databus.InitAegis(nil)
	defer databus.CloseAegis()
	svr := like.New(cfg)
	grpcSvr := grpc.New(nil, svr, cfg)
	http.Init(cfg)
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("activity get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.Background())
			http.CloseService()
			log.Info("activity exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}
