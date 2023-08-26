package service

import (
	"flag"
	"path/filepath"
	"time"

	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/conf"
)

var (
	svr *Service
)

func init() {
	dir, _ := filepath.Abs("../cmd/esports-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	component.InitComponents()
	svr = New(conf.Conf)
	time.Sleep(time.Second)

}

func WithService(f func(s *Service)) func() {
	return func() {
		f(svr)
	}
}
