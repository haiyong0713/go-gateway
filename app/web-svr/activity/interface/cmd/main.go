package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus/report"
	"go-common/library/rate/limit/quota"

	"go-gateway/app/web-svr/activity/interface/rewards"
	"go-gateway/app/web-svr/activity/interface/service/wishes_2021_spring"
	"go-gateway/app/web-svr/activity/interface/tool"

	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/http"
	rpc "go-gateway/app/web-svr/activity/interface/rpc/server"
	"go-gateway/app/web-svr/activity/interface/server/grpc"
	"go-gateway/app/web-svr/activity/interface/service"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	initialize.Init()
	client.New(conf.Conf)
	// init log
	log.Init(conf.Conf.Log)
	trace.Init(conf.Conf.Tracer)
	quota.Init()
	defer func() {
		quota.Close()
		_ = log.Close()
		_ = trace.Close()
	}()
	log.Info("activity start")
	// ecode
	ecode.Init(conf.Conf.Ecode)
	// report
	report.InitUser(nil)

	if err := component.InitByCfg(conf.Conf); err != nil {
		log.Error("sql.InitByCfg error(%v)", err)
		panic(err)
	}

	initialize.Call(component.InitDWRelations)

	initialize.Call(component.InitClient)

	if err := component.InitProducer(conf.Conf); err != nil {
		log.Error("component.InitProducer error(%v)", err)
		panic(err)
	}

	// service init
	service.New(conf.Conf)
	rpcSvr := rpc.New(conf.Conf, service.LikeSvc)
	wishes_2021_spring.InitManuScriptActivityMap(conf.Conf.ManuScriptConfMap)
	http.Init(conf.Conf)
	var grpcSvr *warden.Server
	if !tool.IsBnj2021LiveApplication() {
		grpcSvr = grpc.New(nil, conf.Conf.RpcLimiter)
	}
	// rewards
	rewards.Init(conf.Conf)

	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("activity get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			rpcSvr.Close()
			if grpcSvr != nil {
				_ = grpcSvr.Shutdown(context.Background())
			}
			service.Close()
			component.Close()
			log.Info("activity exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}
