package service

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"go-gateway/app/app-svr/kvo/interface/conf"

	"go-common/library/net/trace"
)

var (
	svr *Service
)

func TestMain(m *testing.M) {
	var (
		err error
	)
	trace.Init(conf.Conf.Tracer)
	defer trace.Close()
	dir, _ := filepath.Abs("../cmd/kvo-test.toml")
	if err = flag.Set("conf", dir); err != nil {
		panic(err)
	}
	if err = conf.Init(); err != nil {
		panic(err)
	}
	svr = New(conf.Conf)
	os.Exit(m.Run())
}
