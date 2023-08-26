package rank

import (
	"flag"
	"path/filepath"
	"time"

	"go-gateway/app/web-svr/activity/job/conf"
)

var svf *Service

func init() {
	dir, _ := filepath.Abs("../../cmd/activity-job-test.toml")
	flag.Set("conf", dir)
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if svf == nil {
		svf = New(conf.Conf)
	}
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(svf)
	}
}
