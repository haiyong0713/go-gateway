package dynamic

import (
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
)

var s *Service

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-dynamic")
		flag.Set("conf_token", "904b98a0103c506237844db17fb61d45")
		flag.Set("tree_id", "159444")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		flag.Set("conf", "../../cmd/app-dynamic-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	infoc := conf.GetInfoc(conf.Conf)
	s = New(conf.Conf, infoc)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}
