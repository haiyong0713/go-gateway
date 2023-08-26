package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-common/component/tinker"
	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/trace"
	"go-common/library/text/translate/chinese.v2"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-resource/interface/component"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/grpc"
	"go-gateway/app/app-svr/app-resource/interface/http"
	"go-gateway/app/app-svr/app-resource/interface/service/abtest"
	absvr "go-gateway/app/app-svr/app-resource/interface/service/abtest"
	auditsvr "go-gateway/app/app-svr/app-resource/interface/service/audit"
	broadcastsvr "go-gateway/app/app-svr/app-resource/interface/service/broadcast"
	deeplinksvr "go-gateway/app/app-svr/app-resource/interface/service/deeplink"
	displaysvr "go-gateway/app/app-svr/app-resource/interface/service/display"
	domainsvr "go-gateway/app/app-svr/app-resource/interface/service/domain"
	entrancesvr "go-gateway/app/app-svr/app-resource/interface/service/entrance"
	fksvr "go-gateway/app/app-svr/app-resource/interface/service/fawkes"
	fpsvr "go-gateway/app/app-svr/app-resource/interface/service/fingerprint"
	fissisvr "go-gateway/app/app-svr/app-resource/interface/service/fission"
	guidesvc "go-gateway/app/app-svr/app-resource/interface/service/guide"
	locsvr "go-gateway/app/app-svr/app-resource/interface/service/location"
	"go-gateway/app/app-svr/app-resource/interface/service/mod"
	"go-gateway/app/app-svr/app-resource/interface/service/notice"
	"go-gateway/app/app-svr/app-resource/interface/service/param"
	pingsvr "go-gateway/app/app-svr/app-resource/interface/service/ping"
	pluginsvr "go-gateway/app/app-svr/app-resource/interface/service/plugin"
	privacysvr "go-gateway/app/app-svr/app-resource/interface/service/privacy"
	showsvr "go-gateway/app/app-svr/app-resource/interface/service/show"
	sidesvr "go-gateway/app/app-svr/app-resource/interface/service/sidebar"
	"go-gateway/app/app-svr/app-resource/interface/service/splash"
	staticsvr "go-gateway/app/app-svr/app-resource/interface/service/static"
	"go-gateway/app/app-svr/app-resource/interface/service/version"
	whitesvr "go-gateway/app/app-svr/app-resource/interface/service/white"
	"go-gateway/app/app-svr/app-resource/interface/service/widget"
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
	cfg := &conf.Config{}
	if err = paladin.Get("app-resource.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("app-resource.toml", cfg); err != nil {
		panic(err)
	}
	abtest.Init(cfg)
	// init log
	log.Init(cfg.Log)
	defer log.Close()
	if err := chinese.NewOpenCC(nil); err != nil {
		log.Error("new opencc error(%v)", err)
		panic(err)
	}
	log.Info("app-resource start")
	// init trace
	trace.Init(cfg.Tracer)
	defer trace.Close()
	// ecode init
	ecode.Init(cfg.Ecode)
	if err := component.InitByCfg(cfg.MySQL.Show); err != nil {
		log.Error("sql.InitByCfg error(%v)", err)
		panic(err)
	}
	ic, _ := infoc.New(cfg.SplashInfoc)
	defer ic.Close()
	anticrawler.Init(nil)
	abt := tinker.Init(ic, &tinker.Config{
		Interval: xtime.Duration(time.Minute),
	})
	defer abt.Close()
	// service init
	svr := initService(cfg, ic)
	http.Init(cfg, svr)
	grpcSvr, err := grpc.New(nil, svr)
	if err != nil {
		panic(err)
	}
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-resource get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			//nolint:errcheck
			grpcSvr.Shutdown(context.TODO())
			log.Info("app-resource exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}

// initService init services.
func initService(c *conf.Config, ic infoc.Infoc) (svr *http.Server) {
	svr = &http.Server{
		AuthSvc: auth.New(nil),
		// init self service,
		PgSvr:          pluginsvr.New(c),
		PingSvr:        pingsvr.New(c),
		SideSvr:        sidesvr.New(c),
		VerSvc:         version.New(c),
		ParamSvc:       param.New(c),
		NtcSvc:         notice.New(c),
		SplashSvc:      splash.New(c, ic),
		AuditSvc:       auditsvr.New(c),
		AbSvc:          absvr.New(c),
		ModSvc:         mod.New(c),
		GuideSvc:       guidesvc.New(c),
		StaticSvc:      staticsvr.New(c),
		DomainSvc:      domainsvr.New(c),
		BroadcastSvc:   broadcastsvr.New(c),
		WhiteSvc:       whitesvr.New(c),
		ShowSvc:        showsvr.New(c, ic),
		FingerPrintSvc: fpsvr.New(c),
		LocationSvc:    locsvr.New(c),
		FkSvc:          fksvr.New(c),
		FissionSvc:     fissisvr.New(c),
		VerifySvr:      verify.New(nil),
		PrivacySvc:     privacysvr.New(c),
		DisplaySvc:     displaysvr.New(c),
		DeeplinkSvc:    deeplinksvr.New(c),
		WidgetSvc:      widget.New(c),
		EntranceSvc:    entrancesvr.New(c),
		FeatureSvc:     feature.New(nil),
		Config:         c,
	}
	return
}
