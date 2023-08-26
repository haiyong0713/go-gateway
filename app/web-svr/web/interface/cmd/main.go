package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/aurora"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/rate/limit/quota"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/http"
	"go-gateway/app/web-svr/web/interface/server/grpc"
	"go-gateway/app/web-svr/web/interface/service"

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
	// aurora
	aurora.Init(nil)
	log.Info("web-interface start")
	// ecode
	ecode.Init(conf.Conf.Ecode)
	// quota
	quota.Init()
	defer quota.Close()
	// anticrawler
	anticrawler.Init(conf.Conf.Anticrawler)
	//server init
	svr := service.New(conf.Conf)
	grpcServ := grpc.New(conf.Conf.GRPCCfg, svr)
	http.Init(conf.Conf, svr)
	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("web-interface get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			if err := grpcServ.Shutdown(ctx); err != nil {
				log.Error("%+v", err)
			}
			log.Info("web-interface exit")
			svr.Close()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}
