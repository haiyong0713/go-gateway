package push

import (
	"flag"
	"os"
	"strings"
	"testing"

	gnk "gopkg.in/h2non/gock.v1"

	"go-gateway/app/app-svr/app-job/job/conf"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-job")
		flag.Set("conf_token", "613aae0ddd1cc47a79920d6115cea472")
		flag.Set("tree_id", "2861")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		flag.Set("conf", "../../cmd/app-job-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	d.client.SetTransport(gnk.DefaultTransport)
	m.Run()
	os.Exit(0)
}

func httpMock(method, url string) *gnk.Request {
	r := gnk.New(url)
	r.Method = strings.ToUpper(method)
	return r
}
