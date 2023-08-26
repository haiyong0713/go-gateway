package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/app-svr/ugc-season/service/conf"
	"go-gateway/app/app-svr/ugc-season/service/server/grpc"
	"go-gateway/app/app-svr/ugc-season/service/server/http"
	"go-gateway/app/app-svr/ugc-season/service/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%+v)", err)
		panic(err)
	}
	// init ecode
	ecode.Init(nil)
	// init log
	log.Init(conf.Conf.Log)
	trace.Init(nil)
	defer trace.Close()
	defer log.Close()
	log.Info("ugc-season-service start")
	svr := service.New(conf.Conf)
	grpcSvr, err := grpc.New(nil, svr, conf.Conf)
	if err != nil {
		panic(err)
	}
	http.Init(conf.Conf)
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("ugc-season-service get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.TODO())
			time.Sleep(time.Second * 2)
			svr.Close()
			log.Info("ugc-season-service exit")
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
