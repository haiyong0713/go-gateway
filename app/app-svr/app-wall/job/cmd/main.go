package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/queue/databus/report"
	"go-gateway/app/app-svr/app-wall/job/conf"
	"go-gateway/app/app-svr/app-wall/job/http"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("app-wall-job.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("app-wall-job.toml", cfg); err != nil {
		panic(err)
	}
	// init log
	log.Init(cfg.Log)
	defer log.Close()
	log.Info("app-wall start")
	// init trace
	trace.Init(cfg.Tracer)
	defer trace.Close()
	// ecode init
	ecode.Init(cfg.Ecode)
	// report init
	report.InitUser(cfg.Report)
	// service init
	http.Init(cfg)
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-wall get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			http.Close()
			log.Info("app-wall exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
