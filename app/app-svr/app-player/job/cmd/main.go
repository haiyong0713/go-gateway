package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-player/job/conf"
	"go-gateway/app/app-svr/app-player/job/http"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	log.Init(conf.Conf.XLog)
	defer log.Close()
	log.Info("app-player-job start")
	http.Init(conf.Conf)
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
