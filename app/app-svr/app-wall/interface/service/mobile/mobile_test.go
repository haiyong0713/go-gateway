package mobile

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-wall/interface/conf"

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
	dir, _ := filepath.Abs("../../cmd/app-wall-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	s = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestMobileState(t *testing.T) {
	Convey("Unicom MobileState", t, WithService(func(s *Service) {
		res := s.MobileState(context.TODO(), "", time.Now())
		So(res, ShouldNotBeEmpty)
	}))
}

func TestUserMobileState(t *testing.T) {
	Convey("Unicom UserMobileState", t, WithService(func(s *Service) {
		res := s.UserMobileState(context.TODO(), "", time.Now())
		So(res, ShouldNotBeEmpty)
	}))
}
