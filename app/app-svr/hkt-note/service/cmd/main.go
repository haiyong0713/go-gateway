package main

import (
	"context"
	"flag"
	"go-common/library/rate/limit/quota"
	"os"
	"os/signal"
	"syscall"

	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/app-svr/hkt-note/service/server/grpc"
	"go-gateway/app/app-svr/hkt-note/service/server/http"
	"go-gateway/app/app-svr/hkt-note/service/service/article"
	"go-gateway/app/app-svr/hkt-note/service/service/image"
	"go-gateway/app/app-svr/hkt-note/service/service/note"

	"go-common/library/conf/paladin.v2"
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
	quota.Init()
	log.Warn("hkt-note-service start")
	svr := initService()
	// http init
	http.Init()
	// grpc init
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
		log.Info("hkt-note-service get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if err := grpcSvr.Shutdown(context.TODO()); err != nil {
				log.Error("grpcSrv.Shutdown error(%v)", err)
			}
			svr.NoteSvr.Close()
			svr.ArtSvr.Close()
			quota.Close()
			log.Info("hkt-note-service exit")
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
	svr = &http.Server{
		NoteSvr: note.New(),
		ImgSvr:  image.New(),
		ArtSvr:  article.New(),
	}
	return
}
