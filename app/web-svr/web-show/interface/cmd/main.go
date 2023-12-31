package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/web-show/interface/conf"
	"go-gateway/app/web-svr/web-show/interface/http"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	// init log
	log.Init(conf.Conf.XLog)
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	defer log.Close()
	log.Info("web-show start")
	// service init
	http.Init(conf.Conf)
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("web-show get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			http.CloseService()
			log.Info("web-show exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
