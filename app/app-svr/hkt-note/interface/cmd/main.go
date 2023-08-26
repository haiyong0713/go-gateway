package main

import (
	"flag"
	"go-common/library/conf/env"
	"go-common/library/conf/paladin.v2"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/net/trace"
	"go-gateway/app/app-svr/hkt-note/interface/conf"
	"go-gateway/app/app-svr/hkt-note/interface/server/http"
	"go-gateway/app/app-svr/hkt-note/interface/service/article"
	"go-gateway/app/app-svr/hkt-note/interface/service/image"
	"go-gateway/app/app-svr/hkt-note/interface/service/note"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	ecode.Init(nil)
	// init log
	log.Init(nil)
	defer log.Close()
	trace.Init(nil)
	defer trace.Close()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	svr := initService()
	// http init
	http.Init(svr)
	// init pprof conf.Conf.Perf
	// init signal
	log.Info("hkt-note-interface start")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("hkt-note-interface get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("hkt-note-interface exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
}

// initService init services.
func initService() (svr *http.Server) {
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	infoc, err := infocV2.New(conf.InfocV2)
	if err != nil {
		if env.DeployEnv == env.DeployEnvProd {
			panic(err)
		}
		log.Error("init service infoc err:%+v", err)
	}

	svr = &http.Server{
		NoteSvr: note.New(conf, infoc),
		ImgSvr:  image.New(conf, infoc),
		ArtSvr:  article.New(conf),
	}
	return
}
