package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/net/trace"
	quota2 "go-common/library/rate/limit/quota"
	"go-gateway/app/app-svr/app-gw/peat-moss/internal/di"

	"github.com/getsentry/sentry-go"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	trace.Init(nil)
	quota2.Init()
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	initSentry()
	log.Info("app-gw-grpc-proxy start")
	//nolint:errcheck
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
			log.Info("app-gw exit")
			sentry.Flush(time.Second * 5)
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func initSentry() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn != "" {
		//nolint:errcheck
		sentry.Init(sentry.ClientOptions{
			Dsn: dsn,
		})
		log.Info("initSentry success(%s)", dsn)
	}
}
