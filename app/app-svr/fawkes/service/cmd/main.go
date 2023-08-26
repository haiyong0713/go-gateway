package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"

	"go-common/library/log/infoc.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/http"
	apmSvr "go-gateway/app/app-svr/fawkes/service/service/apm"
	appSvr "go-gateway/app/app-svr/fawkes/service/service/app"
	authSvr "go-gateway/app/app-svr/fawkes/service/service/auth"
	bizapkSvr "go-gateway/app/app-svr/fawkes/service/service/bizapk"
	buglySvr "go-gateway/app/app-svr/fawkes/service/service/bugly"
	busSvr "go-gateway/app/app-svr/fawkes/service/service/business"
	cdSvr "go-gateway/app/app-svr/fawkes/service/service/cd"
	ciSvr "go-gateway/app/app-svr/fawkes/service/service/ci"
	configSvr "go-gateway/app/app-svr/fawkes/service/service/config"
	feedbackSvr "go-gateway/app/app-svr/fawkes/service/service/feedback"
	ffSvr "go-gateway/app/app-svr/fawkes/service/service/ff"
	gitSvr "go-gateway/app/app-svr/fawkes/service/service/gitlab"
	laserSvr "go-gateway/app/app-svr/fawkes/service/service/laser"
	mngSvr "go-gateway/app/app-svr/fawkes/service/service/manager"
	modSvr "go-gateway/app/app-svr/fawkes/service/service/mod"
	mdlSvr "go-gateway/app/app-svr/fawkes/service/service/modules"
	openSvr "go-gateway/app/app-svr/fawkes/service/service/open"
	pingSvr "go-gateway/app/app-svr/fawkes/service/service/ping"
	prometheusSvr "go-gateway/app/app-svr/fawkes/service/service/prometheus"
	statisticsSvr "go-gateway/app/app-svr/fawkes/service/service/statistics"
	tribeSvr "go-gateway/app/app-svr/fawkes/service/service/tribe"
	webContainerSvr "go-gateway/app/app-svr/fawkes/service/service/webcontainer"
	taskSvr "go-gateway/app/app-svr/fawkes/service/task"
	"go-gateway/app/app-svr/fawkes/service/tools/middleware"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	// init log
	log.Init(conf.Conf.XLog)
	defer log.Close()
	log.Info("fawkes-admin start")
	// init infoc
	ic, err := infoc.New(nil)
	if err != nil {
		panic(err)
	}
	conf.Conf.Infoc = ic
	defer ic.Close()
	// init trace
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	// ecode init
	ecode.Init(nil)
	middleware.Init()
	// service init
	ss := initService(conf.Conf)
	http.Init(conf.Conf, ss)
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("fawkes-admin get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("fawkes-admin exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

// initService init services.
func initService(c *conf.Config) (svr *http.Servers) {
	svr = &http.Servers{
		ConfigSvr:       configSvr.New(c),
		FFSvr:           ffSvr.New(c),
		PingSvr:         pingSvr.New(c),
		AppSvr:          appSvr.New(c),
		CDSvr:           cdSvr.New(c),
		CiSvr:           ciSvr.New(c),
		GitSvr:          gitSvr.New(c),
		BusSvr:          busSvr.New(c),
		MngSvr:          mngSvr.New(c),
		ApmSvr:          apmSvr.New(c),
		LaserSvr:        laserSvr.New(c),
		MdlSvr:          mdlSvr.New(c),
		BizapkSvr:       bizapkSvr.New(c),
		ModSvr:          modSvr.New(c),
		StatisticsSvr:   statisticsSvr.New(c),
		PrometheusSvr:   prometheusSvr.New(c),
		FeedbackSvr:     feedbackSvr.New(c),
		BuglySvr:        buglySvr.New(c),
		TribeSvr:        tribeSvr.New(c),
		OpenSvr:         openSvr.New(c),
		TaskSvr:         taskSvr.New(c),
		AuthSvr:         authSvr.New(c),
		WebContainerSvr: webContainerSvr.New(c),
	}
	return
}
