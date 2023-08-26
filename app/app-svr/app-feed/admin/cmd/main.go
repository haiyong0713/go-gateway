package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/queue/databus/actionlog"
	"go-common/library/queue/databus/report"

	_ "go.uber.org/automaxprocs"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	searchRpc "go-gateway/app/app-svr/app-feed/admin/grpc"
	"go-gateway/app/app-svr/app-feed/admin/http"
	"go-gateway/app/app-svr/app-feed/admin/service/pwd_appeal"
	"go-gateway/app/app-svr/app-feed/admin/service/search"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	log.Init(conf.Conf.Log)
	defer log.Close()
	ecode.Init(conf.Conf.Ecode)
	actionlog.InitManager(nil)
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	report.InitManager(conf.Conf.ManagerReport)
	// service init
	searchSvc := search.New(conf.Conf)
	appealSvc := pwd_appeal.NewService(conf.Conf)
	searchRpcSvc := searchRpc.New(conf.Conf.RPCServer, searchSvc, appealSvc)
	http.Init(conf.Conf, searchSvc)
	log.Info("feed-admin start")

	var (
		c = make(chan os.Signal, 1)
	)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("feed-admin get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			//nolint:errcheck
			searchRpcSvc.Shutdown(context.Background())
			log.Info("feed-admin exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
