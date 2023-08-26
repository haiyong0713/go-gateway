package account

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-wall/job/conf"

	gock "gopkg.in/h2non/gock.v1"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-wall-job")
		flag.Set("conf_token", "66c0ecee0431f5fef5e268819c6044b0")
		flag.Set("tree_id", "22084")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestAddVIP(t *testing.T) {
	Convey("add AddVIP", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", d.addVIPURL).Reply(200).JSON(`{
				"code": 0
			}`)
		err := d.AddVIP(ctx(), 1, 1)
		So(err, ShouldBeNil)
	})
}
