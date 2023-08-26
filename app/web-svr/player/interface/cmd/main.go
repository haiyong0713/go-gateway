package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/player/interface/conf"
	"go-gateway/app/web-svr/player/interface/http"
	"go-gateway/app/web-svr/player/interface/service"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	defer conf.Close()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	log.Init(conf.Conf.XLog)
	defer log.Close()
	log.Info("play-interface start")
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	// ecode
	ecode.Init(conf.Conf.Ecode)
	svr := service.New(conf.Conf)
	// service init
	http.Init(conf.Conf, svr)
	// monitor
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("play-interface get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			svr.Close()
			log.Info("play-interface exit")
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
