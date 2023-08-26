package main

import (
	"flag"
	"go-gateway/app/web-svr/activity/job/component/boss"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/library/log"
	"go-common/library/net/trace"

	"go-common/library/rate/limit/quota"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao"
	"go-gateway/app/web-svr/activity/job/http"
	"go-gateway/app/web-svr/activity/job/service"
	"go-gateway/app/web-svr/activity/job/tool"

	_ "go.uber.org/automaxprocs"
)

var (
	srv *service.Service
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}

	initialize.Init()

	if err := client.InitClients(conf.Conf); err != nil {
		panic(err)
	}

	if err := component.InitByCfg(conf.Conf); err != nil {
		panic(err)
	}
	defer component.Close()

	tool.UpdateCropWeChat(conf.Conf)
	// init log
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	log.Init(conf.Conf.Log)
	defer log.Close()
	quota.Init()
	defer quota.Close()
	dao.New(conf.Conf)
	log.Info("activity-job start")
	boss.NewClient(conf.Conf.Boss)
	srv = service.New(conf.Conf)
	http.Init(conf.Conf, srv)
	signalHandler()
}

func signalHandler() {
	var (
		err error
		ch  = make(chan os.Signal, 1)
	)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("get a signal %s, stop the consume process", si.String())
			if err = srv.Close(); err != nil {
				log.Error("srv close consumer error(%v)", err)
			}
			dao.Close()
			time.Sleep(5 * time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
