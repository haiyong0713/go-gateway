package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/playlist/job/conf"
	"go-gateway/app/web-svr/playlist/job/http"
	"go-gateway/app/web-svr/playlist/job/service"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	initConf()
	log.Init(conf.Conf.Log)
	defer log.Close()
	log.Info("playlist-job start")
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	srv := service.New(conf.Conf)
	http.Init(conf.Conf, srv)
	initSignal(srv)
}

func initConf() {
	if err := conf.Init(); err != nil {
		panic(err)
	}
}

func initSignal(srv *service.Service) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("playlist-job get a signal: %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if err := srv.Close(); err != nil {
				log.Error("srv close consumer error(%v)", err)
			}
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
			// TODO: reload
		default:
			return
		}
	}
}
