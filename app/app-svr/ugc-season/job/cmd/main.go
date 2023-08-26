package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"

	"go-gateway/app/app-svr/ugc-season/job/conf"
	"go-gateway/app/app-svr/ugc-season/job/http"
	"go-gateway/app/app-svr/ugc-season/job/service"

	_ "go.uber.org/automaxprocs"
)

var (
	srv *service.Service
)

func main() {
	flag.Parse()

	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("ugc-season-job.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("ugc-season-job.toml", cfg); err != nil {
		panic(err)
	}

	// init log
	log.Init(cfg.Log)
	defer log.Close()
	log.Info("usc-season-job start")
	srv = service.New(cfg)
	http.Init(cfg, srv)
	signalHandler()
}

func signalHandler() {
	var (
		err error
		ch  = make(chan os.Signal, 1)
	)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("get a signal %s, stop the consume process", si.String())
			if err = srv.Close(); err != nil {
				log.Error("srv close consumer error(%v)", err)
			}
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
