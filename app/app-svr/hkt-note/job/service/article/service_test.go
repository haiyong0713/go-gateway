package article

import (
	"flag"
	"path/filepath"
	"time"

	"go-common/library/net/trace"
	"go-gateway/app/app-svr/hkt-note/job/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	srv *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/hkt-note-job.toml")
	flag.Set("conf", dir)
	conf.Init()
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
