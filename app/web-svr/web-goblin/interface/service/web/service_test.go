package web

import (
	"flag"
	"path/filepath"
	"time"

	"go-gateway/app/web-svr/web-goblin/interface/conf"
)

var svf *Service

func init() {
	dir, _ := filepath.Abs("../cmd/web-goblin-test.toml")
	flag.Set("conf", dir)
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if svf == nil {
		svf = New(conf.Conf)
	}
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(svf)
	}
}
