package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"

	"go-common/library/aurora"
	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/text/translate/chinese.v2"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/grpc"
	"go-gateway/app/app-svr/app-view/interface/http"
	"go-gateway/app/app-svr/app-view/interface/service/report"
	"go-gateway/app/app-svr/app-view/interface/service/view"
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
	if err = paladin.Get("app-view.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("app-view.toml", cfg); err != nil {
		panic(err)
	}
	if err = writeFile("/data/conf/"); err != nil {
		panic(err)
	}
	// init log
	log.Init(cfg.XLog)
	defer log.Close()
	// init chinese
	if err := chinese.NewOpenCC(nil); err != nil {
		log.Error("new opencc error(%v)", err)
		panic(err)
	}
	log.Info("app-view start")
	aurora.Init(nil)
	// init trace
	trace.Init(cfg.Tracer)
	defer trace.Close()
	// ecode init
	ecode.Init(cfg.Ecode)
	// mogul init
	anticrawler.Init(nil)
	svr := initService(cfg)
	// service init
	http.Init(cfg, svr)
	// grpc init
	grpcSvr, err := grpc.New(cfg.RpcServer, svr, cfg)
	if err != nil {
		panic(err)
	}
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-view get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			_ = grpcSvr.Shutdown(context.TODO())
			svr.ViewSvr.Close()
			log.Info("app-view exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

// initService init services.
func initService(c *conf.Config) (svr *http.Server) {
	svr = &http.Server{
		VerifySvc: verify.New(nil),
		AuthSvr:   auth.New(nil),
		ViewSvr:   view.New(c),
		ReportSvr: report.New(c),
		// databus
		UserActPub: databus.New(c.UseractPub),
		DislikePub: databus.New(c.DislikePub),
		FeatureSvc: feature.New(nil),
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
