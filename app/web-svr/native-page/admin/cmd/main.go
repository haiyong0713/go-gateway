package main

import (
	"flag"
	"os"
	"time"

	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/os/signal"
	"go-common/library/syscall"
	_ "go.uber.org/automaxprocs"

	"go-gateway/app/web-svr/native-page/admin/conf"
	"go-gateway/app/web-svr/native-page/admin/http"
	"go-gateway/app/web-svr/native-page/admin/service"
)

var (
	s *service.Service
)

func main() {
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("native-page-admin.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("native-page-admin.toml", cfg); err != nil {
		panic(err)
	}
	log.Init(cfg.Log)
	defer log.Close()
	trace.Init(cfg.Tracer)
	defer trace.Close()
	ecode.Init(nil)
	s = service.New(cfg)
	http.Init(cfg, s)
	log.Info("native-page-admin start")
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
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			time.Sleep(time.Second * 2)
			log.Info("get a signal %s, stop the push-admin process", si.String())
			s.Close()
			s.Wait()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
