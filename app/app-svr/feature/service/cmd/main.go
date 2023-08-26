package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-gateway/app/app-svr/feature/service/conf"
	"go-gateway/app/app-svr/feature/service/http"
	"go-gateway/app/app-svr/feature/service/server/grpc"
	"go-gateway/app/app-svr/feature/service/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	ecode.Init(nil)
	// init log
	log.Init(conf.Conf.Log)
	defer log.Close()
	log.Info("feature-service start")
	// service init
	svr := service.New(conf.Conf)
	// http init
	http.Init(conf.Conf, svr)
	// grpc init
	grpcSvr, err := grpc.New(nil, svr, conf.Conf)
	if err != nil {
		panic(err)
	}
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("feature-service get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.TODO())
			log.Info("feature-service exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}
