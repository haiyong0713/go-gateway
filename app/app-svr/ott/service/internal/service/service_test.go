package service

import (
	"flag"
	"go-gateway/app/app-svr/ott/service/conf"
	"path/filepath"
	"time"
)

var (
	srv *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/ott-service.toml")
	flag.Set("conf", dir)
	conf.Init()
	srv = New(conf.Conf)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(srv)
	}
}
