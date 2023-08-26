package main

import (
	"flag"
	"os"
	"time"

	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/os/signal"
	"go-common/library/queue/databus/report"
	"go-common/library/syscall"
	"go-gateway/app/web-svr/appstatic/admin/conf"
	"go-gateway/app/web-svr/appstatic/admin/http"
	"go-gateway/app/web-svr/appstatic/admin/service"

	"go-common/library/conf/paladin.v2"

	_ "go.uber.org/automaxprocs"
)

var (
	s *service.Service
)

func main() {
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("appstatic-admin.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("appstatic-admin.toml", cfg); err != nil {
		panic(err)
	}
	log.Init(cfg.XLog)
	defer log.Close()
	trace.Init(cfg.Tracer)
	defer trace.Close()
	report.InitManager(nil)
	// service init
	s = service.New(cfg)
	http.Init(cfg, s)
	log.Info("appstatic-admin start")
	signalHandler()
}

func signalHandler() {
	var (
		ch = make(chan os.Signal, 1)
	)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			time.Sleep(time.Second * 2)
			log.Info("get a signal %s, stop the appstatic-admin process", si.String())
			s.Wait()
			s.Close()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
