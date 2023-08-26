package main

import (
	"flag"
	"go-common/library/conf/paladin.v2"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/app-svr/hkt-note/job/http"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	log.Init(nil)
	defer log.Close()
	trace.Init(nil)
	defer trace.Close()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	log.Info("hkt-note-job start")
	http.Init()
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
			http.NoteSrv.Close()
			http.ArtSrv.Close()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
