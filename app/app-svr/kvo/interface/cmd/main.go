package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-gateway/app/app-svr/kvo/interface/conf"
	"go-gateway/app/app-svr/kvo/interface/http"
	"go-gateway/app/app-svr/kvo/interface/server/grpc"
	"go-gateway/app/app-svr/kvo/interface/service"

	"go-common/library/log"
	"go-common/library/net/trace"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	// init log
	log.Init(conf.Conf.XLog)
	defer log.Close()
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()

	log.Info("kvo start")
	svc := service.New(conf.Conf)
	// service init
	http.Init(conf.Conf, svc)
	grpcSvc := grpc.New(conf.Conf.GRPC, svc)

	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("kvo get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("kvo exit")
			grpcSvc.Shutdown(context.TODO())
			svc.Close()
			time.Sleep(1 * time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
