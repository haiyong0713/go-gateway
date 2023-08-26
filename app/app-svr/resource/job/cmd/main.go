package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/job/conf"
	"go-gateway/app/app-svr/resource/job/http"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("resource-job.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("resource-job.toml", cfg); err != nil {
		panic(err)
	}
	log.Init(cfg.Log)
	defer log.Close()
	log.Info("resource-job start")
	http.Init(cfg)
	signalHandler()
}

func signalHandler() {
	var (
		ch = make(chan os.Signal, 1)
	)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("get a signal %s, stop the consume process", si.String())
			http.Svc.Close()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
