package fm

import (
	"flag"
	"path/filepath"
	"time"

	"go-common/library/net/trace"
	"go-gateway/app/app-svr/app-car/job/conf"
)

var (
	srv *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/app-car-job.toml")
	flag.Set("conf", dir)
	if err := conf.Init(); err != nil {
		panic(err)
	}
	trace.Init(conf.Conf.Tracer)
	srv = New(conf.Conf)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(srv)
	}
}
