package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"go-common/component/tinker"
	"go-common/library/aurora"
	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/net/trace"
	"go-common/library/text/translate/chinese.v2"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/http"
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
	if err = paladin.Get("app-feed.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err = paladin.Watch("app-feed.toml", cfg); err != nil {
		panic(err)
	}
	if err = writeFile("/data/conf/"); err != nil {
		panic(err)
	}

	// init log
	log.Init(nil)
	defer log.Close()
	if err := chinese.NewOpenCC(nil); err != nil {
		log.Error("new opencc error(%v)", err)
		panic(err)
	}
	log.Info("app-feed start")
	// init trace
	trace.Init(cfg.Tracer)
	defer trace.Close()
	// ecode init
	ecode.Init(cfg.Ecode)
	anticrawler.Init(nil)
	aurora.Init(nil)

	ic, err := infoc.New(cfg.InfocV2)
	if err != nil {
		panic(err)
	}
	defer ic.Close()
	abt := tinker.Init(ic, &tinker.Config{
		Interval: xtime.Duration(time.Minute),
	})
	defer abt.Close()
	// service init
	http.Init(cfg, ic)
	// init pprof conf.Conf.Perf
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("app-feed get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			http.Close()
			log.Info("app-feed exit")
			return
		case syscall.SIGHUP:
		// TODO reload
		default:
			return
		}
	}
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
