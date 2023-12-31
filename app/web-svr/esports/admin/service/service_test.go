package service

import (
	"flag"
	"path/filepath"
	"time"

	"go-gateway/app/web-svr/esports/admin/conf"
)

var svf *Service

func WithService(f func(s *Service)) func() {
	return func() {
		dir, _ := filepath.Abs("../cmd/esports-admin-test.toml")
		flag.Set("conf", dir)
		conf.Init()
		if svf == nil {
			svf = New(conf.Conf)
		}
		time.Sleep(2 * time.Second)
		f(svf)
	}
}
