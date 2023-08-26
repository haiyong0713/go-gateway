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

	"go-gateway/app/app-svr/ott/service/conf"
	"go-gateway/app/app-svr/ott/service/internal/server/grpc"
	"go-gateway/app/app-svr/ott/service/internal/server/http"
	"go-gateway/app/app-svr/ott/service/internal/service"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("ott service start")
	trace.Init(nil)
	defer trace.Close()
	svc := service.New(conf.Conf)
	grpcSrv := grpc.New(nil, svc)
	httpSrv := http.New(conf.Conf, svc)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			grpcSrv.Shutdown(ctx)
			httpSrv.Shutdown(ctx)
			log.Info("service exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
