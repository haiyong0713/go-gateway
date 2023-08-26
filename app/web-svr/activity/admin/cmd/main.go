package main

import (
	"flag"
	"go-gateway/app/web-svr/activity/admin/client"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"os"
	"time"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/os/signal"
	"go-common/library/syscall"
	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/component/boss"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/http"
	"go-gateway/app/web-svr/activity/admin/service"
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
	initialize.Init()
	log.Init(conf.Conf.Log)
	defer log.Close()
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	ecode.Init(nil)
	component.New(conf.Conf)
	if err := component.InitByCfg(conf.Conf.ORM); err != nil {
		log.Error("sql.InitByCfg error(%v)", err)
		panic(err)
	}
	boss.NewClient(conf.Conf.Boss)
	s = service.New(conf.Conf)
	client.InitClients(conf.Conf)
	http.Init(conf.Conf, s)
	log.Info("activity-admin start")
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
			component.Close()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
