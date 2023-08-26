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
	"go-gateway/app/app-svr/distribution/distribution/internal/di"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto/prelude"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/storagedriver/experimentalflag"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("distribution start")
	trace.Init(nil)
	defer trace.Close()
	//nolint:errcheck
	paladin.Init()
	extension.Init()
	_, closeFunc, err := di.InitApp()
	if err != nil {
		log.Error("Failed to init app: %+v", err)
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
			log.Info("distribution exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
