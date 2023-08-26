package main

import (
	"flag"
	ecode "go-common/library/ecode/tip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/queue/databus/actionlog"

	"go-gateway/app/app-svr/archive-push/admin/internal/di"
	qqDAO "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/dao"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()

	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("archive-push-admin start")
	trace.Init(nil)
	defer trace.Close()
	actionlog.InitManager(nil)
	ecode.Init(nil)
	paladin.Init()
	qqDAO.Init()
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
			log.Info("archive-push-admin exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
