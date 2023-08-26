package mod

import (
	"time"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

var srv *Service

func init() {
	err := conf.Init()
	if err != nil {
		panic(err)
	}
	srv = New(conf.Conf)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(srv)
	}
}
