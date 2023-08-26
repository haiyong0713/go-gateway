package delay

import (
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin"
	"go-common/library/net/trace"
	"go-common/library/testing/dockertest"
)

var testDao *dao

func TestMain(m *testing.M) {
	flag.Set("conf", "./configs")
	flag.Parse()
	trace.Init(nil)
	dockertest.Run("./test/docker-compose.yaml")
	var err error
	if err = paladin.Init(); err != nil {
		panic(err)
	}
	var cfg Cfg
	if err = paladin.Get("delay.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if testDao, err = NewDao(cfg); err != nil {
		panic(err)
	}
	ret := m.Run()
	testDao.Close()
	os.Exit(ret)
}
