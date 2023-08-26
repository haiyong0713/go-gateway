package rank

import (
	"flag"
	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"path/filepath"
	"time"
)

var (
	s *Service
)

func WithService(f func(s *Service)) func() {
	return func() {
		f(s)
	}
}

func init() {
	dir, _ := filepath.Abs("../../cmd/app-show-test.toml")
	flag.Set("conf", dir)
	cfg, err := confInit()
	if err != nil {
		panic(err)
	}
	s = New(cfg)
	time.Sleep(time.Second)
}

func confInit() (*conf.Config, error) {
	err := paladin.Init()
	if err != nil {
		return nil, err
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("app-show.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
