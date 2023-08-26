package daily

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model"

	. "github.com/smartystreets/goconvey/convey"
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

func TestDaily(t *testing.T) {
	Convey("get Daily data", t, WithService(func(s *Service) {
		res := s.Daily(context.TODO(), model.PlatIPhone, 100000, 4, 1, 20)
		So(res, ShouldNotBeEmpty)
	}))
}
