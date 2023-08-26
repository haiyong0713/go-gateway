package service

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/web-svr/native-page/admin/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var svr *Service

func init() {
	dir, _ := filepath.Abs("../cmd/activity-admin-test.toml")
	flag.Set("conf", dir)
	cfg := &conf.Config{}
	if err := paladin.Get("native-page-admin.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	svr = New(cfg)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		Reset(func() {})
		f(svr)
	}
}

func Test_Service(t *testing.T) {
	Convey("service test", t, WithService(func(s *Service) {
		s.Wait()
		s.Close()
	}))
}
