package service

import (
	"flag"
	"path/filepath"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/conf"
)

var svf *Service

func WithService(f func(s *Service)) func() {
	return func() {
		dir, _ := filepath.Abs("../cmd/space-test.toml")
		flag.Set("conf", dir)
		conf.Init()
		log.Init(conf.Conf.Log)
		svf = New(conf.Conf)
		f(svf)
	}
}
