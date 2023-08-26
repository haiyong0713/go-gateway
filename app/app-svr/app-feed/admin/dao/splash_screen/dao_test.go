package splash_screen

import (
	"flag"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"os"
	"testing"
)

var testD *Dao

func TestMain(m *testing.M) {
	flag.Set("conf", "../../cmd/feed-admin-test.toml")
	flag.Set("deploy_env", "uat")
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	}
	testD = New(conf.Conf)
	ret := m.Run()
	os.Exit(ret)
}
