package alarm

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-resource")
		flag.Set("conf_token", "z8JNX5MFIyDxyBsqwQyF6pnjWQ5YOA14")
		flag.Set("tree_id", "2722")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-resource-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestSendWeChart(t *testing.T) {
	convey.Convey("SendWeChart", t, func() {
		convey.Convey("When everything is correct", func(ctx convey.C) {
			httpMock("POST", "http://bap.bilibili.co/api/v1/message/add").Reply(200).JSON("{}")
			err := d.SendWeChart(context.Background(), "", []string{})
			ctx.Convey("Then err should be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		convey.Convey("When set http request gets 404", func(ctx convey.C) {
			httpMock("POST", "http://bap.bilibili.co/api/v1/message/add").Reply(404)
			err := d.SendWeChart(context.Background(), "", []string{})
			ctx.Convey("Then err should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}
