package main

import (
	"flag"
	"os"
	"time"

	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/os/signal"
	report "go-common/library/queue/databus/actionlog"
	"go-common/library/syscall"
	"go-gateway/app/web-svr/space/admin/conf"
	"go-gateway/app/web-svr/space/admin/http"
	"go-gateway/app/web-svr/space/admin/service"
	_ "go.uber.org/automaxprocs"
)

var (
	s *service.Service
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	log.Init(conf.Conf.Log)
	defer log.Close()
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	// service init
	s = service.New(conf.Conf)
	http.Init(conf.Conf, s)
	report.InitManager(conf.Conf.ManagerReport)
	log.Info("space-admin start")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			log.Info("get a signal %s, stop the space-admin process", si.String())
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
