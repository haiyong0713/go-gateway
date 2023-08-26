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
	"go-common/library/rate/limit/quota"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"

	"go-gateway/app/web-svr/esports/interface/client"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/http"
	"go-gateway/app/web-svr/esports/interface/server/grpc"
	"go-gateway/app/web-svr/esports/interface/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init error(%v)", err)
		panic(err)
	}
	initialize.Init()
	if err := client.New(conf.Conf); err != nil {
		panic(err)
	}

	quota.Init()
	log.Init(conf.Conf.Log)
	trace.Init(conf.Conf.Tracer)
	defer func() {
		_ = trace.Close()
		_ = log.Close()
		quota.Close()
	}()
	log.Info("esports start")
	// ecode
	ecode.Init(conf.Conf.Ecode)
	initialize.Call(component.InitComponents)
	initialize.Call(component.InitClient)
	initialize.Call(component.InitProducer)
	//server init
	svr := service.New(conf.Conf)
	grpcSvr := grpc.New(nil, svr, conf.Conf.RpcLimiter)
	http.Init(conf.Conf, svr)
	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("esports get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			log.Info("esports exit")
			grpcSvr.Shutdown(context.Background())
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
