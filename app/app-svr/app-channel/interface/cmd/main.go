package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/text/translate/chinese.v2"
	"go-gateway/app/app-svr/app-channel/interface/conf"
	"go-gateway/app/app-svr/app-channel/interface/http"
	wardensdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	var (
		cfg              = conf.Conf
		sdkBuilderConfig = wardensdk.SDKBuilderConfig{}
	)
	if err = paladin.Get("grpc-client-sdk.toml").UnmarshalTOML(&sdkBuilderConfig); err != nil {
		panic(err)
	}
	conf.InitWardenSDKBuilder(sdkBuilderConfig)
	if err = paladin.Get("app-channel.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("app-channel.toml", cfg); err != nil {
		panic(err)
	}
	ecode.Init(nil)
	// init log
	log.Init(conf.Conf.Log)
	defer log.Close()
	if err := chinese.NewOpenCC(nil); err != nil {
		log.Error("new opencc error(%v)", err)
		panic(err)
	}
	log.Info("app-channel start")
	// init trace
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	// service init
	http.Init(conf.Conf)
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-channel get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("app-channel exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}
