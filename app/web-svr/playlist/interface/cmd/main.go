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
	"go-gateway/app/web-svr/playlist/interface/conf"
	"go-gateway/app/web-svr/playlist/interface/http"
	rpc "go-gateway/app/web-svr/playlist/interface/rpc/server"
	"go-gateway/app/web-svr/playlist/interface/server/grpc"
	"go-gateway/app/web-svr/playlist/interface/service"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init error(%v)", err)
		panic(err)
	}
	log.Init(conf.Conf.Log)
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	defer log.Close()
	log.Info("playlist start")
	// ecode
	ecode.Init(conf.Conf.Ecode)
	//server init
	svr := service.New(conf.Conf)
	rpcSvr := rpc.New(conf.Conf, svr)
	grpcSvr := grpc.New(nil, svr)
	http.Init(conf.Conf, svr)
	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("playlist get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			grpcSvr.Shutdown(context.Background())
			rpcSvr.Close()
			log.Info("playlist exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
