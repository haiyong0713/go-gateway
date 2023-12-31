package ping

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

func TestPing(t *testing.T) {
	Convey("Unicom Ping", t, WithService(func(s *Service) {
		err := s.Ping(context.TODO())
		So(err, ShouldBeNil)
	}))
}
