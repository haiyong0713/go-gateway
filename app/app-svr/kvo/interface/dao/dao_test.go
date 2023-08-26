package dao

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"go-gateway/app/app-svr/kvo/interface/conf"
)

var (
	testDao *Dao
	c       = context.TODO()
)

func TestMain(m *testing.M) {
	var (
		err error
	)
	dir, _ := filepath.Abs("../cmd/kvo-test.toml")
	if err = flag.Set("conf", dir); err != nil {
		panic(err)
	}
	if err = conf.Init(); err != nil {
		panic(err)
	}
	testDao = New(conf.Conf)
	os.Exit(m.Run())
}
