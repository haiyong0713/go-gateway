package note

import (
	"flag"
	"go-common/library/conf/env"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/net/trace"
	"path/filepath"
	"time"

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
	infoc, err := infocV2.New(conf.InfocV2)
	if err != nil {
		if env.DeployEnv == env.DeployEnvProd {
			panic(err)
		}
		log.Error("init service infoc err:%+v", err)
	}
	srv = New(conf, infoc)
	trace.Init(nil)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		Reset(func() {})
		f(srv)
	}
}
