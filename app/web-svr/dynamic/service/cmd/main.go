package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/dynamic/service/conf"
	"go-gateway/app/web-svr/dynamic/service/http"
	rpc "go-gateway/app/web-svr/dynamic/service/rpc/server"
	"go-gateway/app/web-svr/dynamic/service/server/grpc"
	"go-gateway/app/web-svr/dynamic/service/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	defer conf.Close()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	log.Init(conf.Conf.Log)
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	defer log.Close()
	log.Info("dynamic-service start")
	// service init
	svr := service.New(conf.Conf)
	rpcSvr := rpc.New(conf.Conf, svr)
	gRpc := grpc.New(conf.Conf.GRPC, svr)
	http.Init(conf.Conf, svr)
	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("dynamic-service get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			if err := gRpc.Shutdown(context.Background()); err != nil {
				log.Error("%+v", err)
			}
			rpcSvr.Close()
			svr.Close()
			log.Info("dynamic-service exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}
