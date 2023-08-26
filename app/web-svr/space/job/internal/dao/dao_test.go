package dao

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin"
)

var d *dao
var ctx = context.Background()

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "local" {
		flag.Set("app_id", "main.web-svr.space-job")
		flag.Set("conf_token", "3386d9f263deeda2667a1922cd9e7b4b")
		flag.Set("tree_id", "286584")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../cmd/web-interface-test.toml")
	}
	flag.Parse()
	var err error
	if err = paladin.Init(); err != nil {
		panic(err)
	}
	var cf func()
	if d, cf, err = newTestDao(); err != nil {
		panic(err)
	}
	ret := m.Run()
	cf()
	os.Exit(ret)
}
