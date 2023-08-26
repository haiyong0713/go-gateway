package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"

	"go-gateway/app/web-svr/dance-taiko/job/conf"
	"go-gateway/app/web-svr/dance-taiko/job/http"
	"go-gateway/app/web-svr/dance-taiko/job/service"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Init(conf.Conf.Log)
	defer log.Close()
	log.Info("dance-taiko start")
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	ecode.Init(conf.Conf.Ecode)
	svr := service.New(conf.Conf)
	http.Init(conf.Conf, svr)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("dance-taiko-job get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			svr.Close()
			time.Sleep(time.Second * 2)
			log.Info("dance-taiko-job exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
