package common

import (
	"flag"
	"path/filepath"
	"time"

	"go-common/library/net/trace"
	"go-gateway/app/app-svr/app-car/interface/conf"
)

var (
	srv *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/app-car.toml")
	flag.Set("conf", dir)
	conf.Init()
	trace.Init(conf.Conf.Tracer)
	srv = New(conf.Conf)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(srv)
	}
}
