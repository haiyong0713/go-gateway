package like

import (
	"flag"
	"os"
	"strings"
	"testing"

	gnk "gopkg.in/h2non/gock.v1"

	"go-gateway/app/web-svr/activity/job/conf"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.web-svr.activity-job")
		flag.Set("conf_token", "7c164822b6da4198f6348599bedf1797")
		flag.Set("tree_id", "2703")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/activity-job-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}

func httpMock(method, url string) *gnk.Request {
	r := gnk.New(url)
	r.Method = strings.ToUpper(method)
	d.httpClient.SetTransport(gnk.DefaultTransport)
	return r
}
