package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/conf"
	"go-gateway/app/web-svr/esports/job/http"
	"go-gateway/app/web-svr/esports/job/service"
	"go-gateway/app/web-svr/esports/job/sql"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	initialize.Init()
	log.Init(conf.Conf.Log)
	defer log.Close()
	log.Info("esports start")
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	ecode.Init(conf.Conf.Ecode)
	// service init
	initialize.Call(sql.InitByCfg)
	// grpc init
	initialize.Call(component.InitClients)
	component.InitRedis()
	component.InitCache()
	initialize.Call(component.InitComponents)
	svr := service.NewV2(conf.Conf)
	http.Init(conf.Conf, svr)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("esports get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("esports job exit")
			svr.Close()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
