package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	mauth "go-common/component/auth/middleware/grpc"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"

	infocV2 "go-common/library/log/infoc.v2"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	dynGrpcV1 "go-gateway/app/app-svr/app-dynamic/interface/grpc"
	dynGrpcV2 "go-gateway/app/app-svr/app-dynamic/interface/grpc/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/http"
	drawsvc "go-gateway/app/app-svr/app-dynamic/interface/service/draw"
	dynamicsvc "go-gateway/app/app-svr/app-dynamic/interface/service/dynamic"
	dynamicsvcV2 "go-gateway/app/app-svr/app-dynamic/interface/service/dynamicV2"
	topicsvc "go-gateway/app/app-svr/app-dynamic/interface/service/topic"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	ecode.Init(nil)
	// init log
	log.Init(conf.Conf.Log)
	defer log.Close()
	log.Info("app-dynamic start")
	// init trace
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	anticrawler.Init(nil)
	// infoc v2
	infoc := conf.GetInfoc(conf.Conf)
	// service init
	svr := initServer(conf.Conf, infoc)
	// http init
	http.Init(conf.Conf, svr)
	// grpc init
	grpcSvr, err := initGRPCServer(conf.Conf.MossGRPC, svr)
	if err != nil {
		panic(err)
	}
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-dynamic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.TODO())
			log.Info("app-dynamic exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}

func initServer(c *conf.Config, infoc infocV2.Infoc) *http.Server {
	return &http.Server{
		DynamicSvc:   dynamicsvc.New(c, infoc),
		VerifySvc:    verify.New(nil),
		AuthSvc:      auth.New(nil),
		DrawSvc:      drawsvc.New(c),
		DynamicSvcV2: dynamicsvcV2.New(c, infoc),
		Config:       c,
		TopicSvc:     topicsvc.New(c),
		FeatureSvc:   feature.New(nil),
	}
}

func initGRPCServer(c *warden.ServerConfig, svr *http.Server) (*warden.Server, error) {
	wsvr := warden.NewServer(c)
	authM := mauth.New(nil)
	_, err := dynGrpcV1.New(wsvr, authM, svr)
	if err != nil {
		return nil, err
	}
	_, err = dynGrpcV2.New(wsvr, authM, svr)
	if err != nil {
		return nil, err
	}
	_, err = dynGrpcV2.InitCampusSvr(wsvr, authM, svr)
	if err != nil {
		return nil, err
	}
	wsvr, err = wsvr.Start()
	if err != nil {
		return nil, err
	}
	return wsvr, nil
}
