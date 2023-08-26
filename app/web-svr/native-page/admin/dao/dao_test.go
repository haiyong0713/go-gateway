package dao

import (
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/web-svr/native-page/admin/conf"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.web-svr.activity-admin")
		flag.Set("conf_token", "a0c33c892e0c08476ecbb5d28e5880cf")
		flag.Set("tree_id", "34245")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../cmd/activity-admin-test.toml")
	}
	flag.Parse()
	cfg := &conf.Config{}
	if err := paladin.Get("native-page-admin.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	d = New(cfg)
	os.Exit(m.Run())
}
