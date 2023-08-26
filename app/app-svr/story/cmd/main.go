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
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/net/trace"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/story/internal/di"

	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Parse()
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("story start")
	trace.Init(nil)
	defer trace.Close()
	anticrawler.Init(nil)
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	if err := writeFile("/data/conf/"); err != nil {
		panic(err)
	}
	ic, err := infocV2.New(nil)
	if err != nil {
		panic(err)
	}
	abt := tinker.Init(ic, &tinker.Config{
		Interval: xtime.Duration(time.Minute),
	})
	defer abt.Close()
	_, closeFunc, err := di.InitApp(ic)
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.Info("story exit")
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
