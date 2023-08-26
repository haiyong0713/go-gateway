package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"go-common/library/aurora"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/rate/limit/quota"
	"go-common/library/text/translate/chinese.v2"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-search/internal/di"

	"github.com/pkg/errors"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	log.Init(nil)
	defer log.Close()
	log.Info("app-search start")
	trace.Init(nil)
	defer trace.Close()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	if err := writeFile("/data/conf/"); err != nil {
		panic(err)
	}
	// init chinese
	if err := chinese.NewOpenCC(nil); err != nil {
		log.Error("new opencc error(%v)", err)
		panic(err)
	}
	// init aurora
	aurora.Init(nil)
	anticrawler.Init(nil)
	// quota init
	quota.Init()
	defer quota.Close()
	_, closeFunc, err := di.InitApp()
	if err != nil {
		panic(errors.WithStack(err))
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.Info("app-search exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
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
