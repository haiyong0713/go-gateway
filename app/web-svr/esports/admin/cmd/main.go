package main

import (
	"flag"
	"os"
	"time"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/os/signal"
	"go-common/library/syscall"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"go-gateway/app/web-svr/esports/admin/client"
	"go-gateway/app/web-svr/esports/admin/component"
	"go-gateway/app/web-svr/esports/admin/conf"
	"go-gateway/app/web-svr/esports/admin/http"
	"go-gateway/app/web-svr/esports/admin/service"

	_ "go.uber.org/automaxprocs"
)

var (
	s *service.Service
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
	defer component.Close()
	if err := client.InitClients(conf.Conf); err != nil {
		panic(err)
	}
	// service init
	s = service.New(conf.Conf)
	http.Init(conf.Conf, s)

	log.Info("esports-admin start")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			log.Info("get a signal %s, stop the esports-admin process", si.String())
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
