package main

import (
	"flag"
	"go-gateway/app/app-svr/archive-extra-shjd/job/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"

	_ "go.uber.org/automaxprocs"

	"go-gateway/app/app-svr/archive-extra-shjd/job/conf"
	"go-gateway/app/app-svr/archive-extra-shjd/job/service"
)

func main() {
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()

	cfg := &conf.Config{}
	if err = paladin.Get("archive-extra-job-shjd.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("archive-extra-job-shjd.toml", cfg); err != nil {
		panic(err)
	}

	// init ecode
	ecode.Init(nil)
	// init log
	log.Init(cfg.Log)
	trace.Init(nil)
	defer trace.Close()
	defer log.Close()

	log.Info("archive-extra-job-shjd start")

	srv := service.New(cfg)
	http.Init(cfg, srv)

	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("archive-extra-job-shjd get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			time.Sleep(time.Second * 2)
			log.Info("archive-extra-job-shjd exit")
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
