package unicom

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-wall/job/conf"
	"go-gateway/app/app-svr/app-wall/job/model/unicom"

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
	dir, _ := filepath.Abs("../../cmd/app-wall-job-test.toml")
	flag.Set("conf", dir)
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("app-wall-job.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("app-wall-job.toml", cfg); err != nil {
		panic(err)
	}
	s = New(cfg)
	time.Sleep(time.Second)
}

func TestAddUserIntegralLog(t *testing.T) {
	Convey("Unicom addUserIntegralLog", t, WithService(func(s *Service) {
		s.addUserIntegralLog(&unicom.UserPackLog{})
	}))
}

func TestLoadUnicomIP(t *testing.T) {
	Convey("Unicom loadUnicomIP", t, WithService(func(s *Service) {
		s.loadUnicomIP(context.TODO())
	}))
}

func TestLoadUnicomIPOrder(t *testing.T) {
	Convey("Unicom loadUnicomIPOrder", t, WithService(func(s *Service) {
		s.loadUnicomIPOrder()
	}))
}
