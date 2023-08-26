package dao

import (
	"flag"
	"os"
	"testing"

	"go-gateway/app/web-svr/player/interface/conf"

	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "local" {
		flag.Set("app_id", "main.web-svr.player-interface")
		flag.Set("conf_token", "419a7d1d51888c17a9dbdc0abb61cee8")
		flag.Set("tree_id", "2285")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../cmd/player-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	d.client.SetTransport(gock.DefaultTransport)
	os.Exit(m.Run())
}
