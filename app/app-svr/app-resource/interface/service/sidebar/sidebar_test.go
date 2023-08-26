package sidebar

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-resource/interface/conf"

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
	dir, _ := filepath.Abs("../../cmd/app-resource-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	s = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestSideBar(t *testing.T) {
	Convey("get SideBar data", t, WithService(func(s *Service) {
		res := s.SideBar(context.TODO(), 1, 1, 0, 1, 1, "hans")
		So(res, ShouldNotBeEmpty)
	}))
}
