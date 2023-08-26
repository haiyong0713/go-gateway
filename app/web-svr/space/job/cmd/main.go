package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/railgun.v2"
	"go-gateway/app/web-svr/space/job/internal/di"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("space-job start")
	trace.Init(nil)
	defer trace.Close()
	// 程序启动时初始化railgun
	railgun.Init(nil)
	_ = paladin.Init()
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
			log.Info("space-job exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
