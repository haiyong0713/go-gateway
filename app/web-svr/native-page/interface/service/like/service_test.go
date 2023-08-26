package like

import (
	"flag"
	"path/filepath"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/web-svr/native-page/interface/conf"
)

var svf *Service

func init() {
	dir, _ := filepath.Abs("../../cmd/activity-test.toml")
	flag.Set("conf", dir)
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	cfg := &conf.Config{}
	if err := paladin.Get("native-page-interface.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if svf == nil {
		svf = New(cfg)
	}
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(svf)
	}
}
