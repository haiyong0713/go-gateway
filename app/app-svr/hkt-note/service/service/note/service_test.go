package note

import (
	"flag"
	"path/filepath"
	"time"

	"go-common/library/net/trace"
	"go-gateway/app/app-svr/hkt-note/service/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	srv *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/hkt-note-service.toml")
	flag.Set("conf", dir)
	srv = New()
	trace.Init(conf.Conf.Tracer)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		Reset(func() {})
		f(srv)
	}
}
