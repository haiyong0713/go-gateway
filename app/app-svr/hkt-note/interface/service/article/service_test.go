package article

import (
	"flag"
	"go-common/library/conf/paladin.v2"
	"path/filepath"
	"time"

	"go-common/library/net/trace"
	"go-gateway/app/app-svr/hkt-note/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	srv *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/hkt-note.toml")
	flag.Set("conf", dir)
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	srv = New(conf)
	trace.Init(nil)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		Reset(func() {})
		f(srv)
	}
}
