package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"

	"go-common/component/tinker"
	"go-common/library/aurora"
	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/queue/databus/report"
	"go-common/library/rate/limit/quota"
	"go-common/library/text/translate/chinese.v2"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/media"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/grpc"
	"go-gateway/app/app-svr/app-interface/interface-legacy/http"
	acc "go-gateway/app/app-svr/app-interface/interface-legacy/service/account"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/dataflow"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/display"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/favorite"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/history"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/relation"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/search"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/space"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/teenagers"
	feature "go-gateway/app/app-svr/feature/service/sdk"

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
		cfg = conf.Conf
	)
	if err = paladin.Get("app-interface.toml").UnmarshalTOML(cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("app-interface.toml", cfg); err != nil {
		panic(err)
	}
	if err = writeFile("/data/conf/"); err != nil {
		panic(err)
	}
	// init log
	log.Init(nil)
	defer log.Close()
	log.Info("app-interface start")
	// init chinese
	if err := chinese.NewOpenCC(nil); err != nil {
		log.Error("new opencc error(%v)", err)
		panic(err)
	}
	// init aurora
	aurora.Init(nil)
	// init trace
	trace.Init(cfg.Tracer)
	defer trace.Close()
	anticrawler.Init(nil)
	report.InitUser(cfg.Report)
	// ecode init
	ecode.Init(cfg.Ecode)
	// quota init
	quota.Init()
	defer quota.Close()
	ic, _ := infoc.New(nil)
	cfg.Infocv2 = ic
	defer ic.Close()
	// service init
	svr := initService(cfg)
	// http init
	http.Init(cfg, svr)
	// grpc init
	grpcSvr := grpc.New(cfg.Warden, svr)
	//init tinker
	abt := tinker.Init(ic, nil)
	defer abt.Close()
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-interface get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			//nolint:errcheck
			grpcSvr.Shutdown(context.TODO())
			log.Info("app-interface exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func initService(c *conf.Config) (svr *http.Server) {
	svr = &http.Server{
		VerifySvc:   verify.New(nil),
		AuthSvc:     auth.New(nil),
		SpaceSvr:    space.New(c),
		SrcSvr:      search.New(c),
		DisplaySvr:  display.New(c),
		FavSvr:      favorite.New(c),
		AccSvr:      acc.New(c),
		RelSvr:      relation.New(c),
		HistorySvr:  history.New(c),
		MediaSvr:    media.New(c),
		TeenSvr:     teenagers.New(c),
		DataflowSvr: dataflow.New(c),
		UserActPub:  databus.New(c.UseractPub),
		Config:      c,
		FeatureSvc:  feature.New(nil),
	}
	return
}

// writeFile : Static initialization
//
//nolint:gosec
func writeFile(p string) error {
	for _, key := range paladin.KeysWithFormat() {
		content, err := paladin.Get(key).String()
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path.Join(p, key), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}
