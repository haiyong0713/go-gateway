package note

import (
	"flag"
	"go-common/library/net/trace"
	"path/filepath"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/hkt-note/job/conf"
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
