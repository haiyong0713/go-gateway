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
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/http"
	"go-gateway/app/web-svr/space/interface/server/grpc"
	"go-gateway/app/web-svr/space/interface/service"

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
	log.Info("space-interface start")
	// ecode
	ecode.Init(conf.Conf.Ecode)
	// anticrawler
	anticrawler.Init(conf.Conf.Anticrawler)
	//server init
	svr := service.New(conf.Conf)
	grpcSvr := grpc.New(conf.Conf.GRPCServer, svr)
	http.Init(conf.Conf, svr)
	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("space get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if err := grpcSvr.Shutdown(context.Background()); err != nil {
				log.Error("%+v", err)
			}
			time.Sleep(time.Second)
			if err := svr.Close(); err != nil {
				log.Error("srv close error(%v)", err)
			}
			log.Info("space-interface exit")
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
