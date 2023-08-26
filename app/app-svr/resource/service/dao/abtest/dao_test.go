package abtest

import (
	"flag"
	"os"
	"strings"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/resource/service/conf"

	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.resource-service")
		flag.Set("conf_token", "a1bf4b2063965fbc2345edb9ab11baf8")
		flag.Set("tree_id", "3232")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/resource-service-test.toml")
	}
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("resource-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	d = New(cfg)
	d.httpClient.SetTransport(gock.DefaultTransport)
	m.Run()
	os.Exit(0)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}
