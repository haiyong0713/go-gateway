package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"

	"go-gateway/app/app-svr/steins-gate/admin/conf"
	"go-gateway/app/app-svr/steins-gate/admin/internal/server/http"
	"go-gateway/app/app-svr/steins-gate/admin/internal/service"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	log.Init(conf.Conf.XLog) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("stein_gate-service start")
	trace.Init(nil)
	defer trace.Close()
	ecode.Init(nil)
	svc := service.New(conf.Conf)
	httpSrv := http.New(conf.Conf, svc)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			//nolint:errcheck
			httpSrv.Shutdown(ctx)
			log.Info("stein_gate-service exit")
			svc.Close()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}

}
