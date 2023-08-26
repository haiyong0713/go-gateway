package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/http"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	// init log
	log.Init(conf.Conf.Log)
	defer log.Close()
	log.Info("app-car start")
	// init trace
	trace.Init(nil)
	defer trace.Close()
	// ecode init
	ecode.Init(nil)
	// service init
	http.Init(conf.Conf)
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-car get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			log.Info("app-car exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}
