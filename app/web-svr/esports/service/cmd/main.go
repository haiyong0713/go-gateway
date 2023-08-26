package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"go-gateway/app/web-svr/esports/service/component"
	"go-gateway/app/web-svr/esports/service/conf"
	"go-gateway/app/web-svr/esports/service/internal/server/grpc"
	"go-gateway/app/web-svr/esports/service/internal/server/http"
	"go-gateway/app/web-svr/esports/service/internal/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	initialize.Init()
	ecode.Init(nil)

	log.Init(conf.Conf.Log)
	defer log.Close()
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	initialize.Call(component.InitByCfg)
	svr := service.New(conf.Conf)
	_, err := grpc.New(nil, svr, conf.Conf.RpcLimiter)
	if err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	_, err = http.New(svr)
	if err != nil {
		panic(err)
	}

	log.Info("esports-service start")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("service exit")
			svr.Close()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
