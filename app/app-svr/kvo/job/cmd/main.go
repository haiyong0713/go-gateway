package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-gateway/app/app-svr/kvo/job/internal/di"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/net/trace"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	trace.Init(nil)
	defer trace.Close()
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("job start")
	paladin.Init()
	_, closeFunc, err := di.InitApp()
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.Info("job exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
