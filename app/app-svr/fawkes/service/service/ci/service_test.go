package ci

import (
	"flag"
	"path/filepath"
	"time"

	"go-common/library/net/trace"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

var (
	srv *Service
)

func init() {
	dir, _ := filepath.Abs("../../conf/fawkes-admin.toml")
	flag.Set("conf", dir)
	err := conf.Init()
	if err != nil {
		return
	}
	srv = New(conf.Conf)
	trace.Init(conf.Conf.Tracer)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		Reset(func() {})
		f(srv)
	}
}
