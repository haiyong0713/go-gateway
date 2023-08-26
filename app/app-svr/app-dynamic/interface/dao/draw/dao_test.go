package dao

import (
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/service/draw"

	"go-common/library/log"
)

var (
	d *Dao
	s *draw.Service
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../cmd/app-dynamic-test.toml")
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	} // init log

	d = New(conf.Conf)
	m.Run()
	s = draw.New(conf.Conf)

	os.Exit(0)
}
