package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/component/tinker"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/net/trace"
	"go-common/library/rate/limit/quota"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/topic/interface/internal/di"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("topic start")
	trace.Init(nil)
	defer trace.Close()
	quota.Init()
	defer quota.Close()
	anticrawler.Init(nil)
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	_, closeFunc, err := di.InitApp()
	if err != nil {
		panic(err)
	}
	//init tinker
	ic, _ := infocV2.New(nil)
	defer ic.Close()
	abt := tinker.Init(ic, nil)
	defer abt.Close()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.Info("topic exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
