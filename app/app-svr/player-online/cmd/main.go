package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs"

	"go-common/library/log"
	"go-common/library/net/trace"

	"go-gateway/app/app-svr/player-online/internal/conf"
	"go-gateway/app/app-svr/player-online/internal/server/grpc"
	"go-gateway/app/app-svr/player-online/internal/server/http"
	"go-gateway/app/app-svr/player-online/internal/service/online"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	log.Init(conf.Conf.Xlog)
	defer log.Close()
	log.Info("player-online start")
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()

	// service init
	svr := online.New(conf.Conf)
	grpcSvr, err := grpc.New(nil, svr, conf.Conf)
	http.Init(conf.Conf, svr)

	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.TODO())
			time.Sleep(time.Second * 2)
			log.Info("player-online exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
