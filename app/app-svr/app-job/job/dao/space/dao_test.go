package space

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-job/job/conf"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var d *Dao

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
	} else {
		flag.Set("conf", "../../cmd/app-job-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	d.client.SetTransport(gock.DefaultTransport)
	d.clientAsyn.SetTransport(gock.DefaultTransport)
	m.Run()
	os.Exit(0)
}
func Test_UpArchives(t *testing.T) {
	Convey("UpArchives", t, func() {
		_, err := d.UpArchives(context.TODO(), 1, 1, 10, "")
		So(err, ShouldBeNil)
	})
}

func Test_UpArticles(t *testing.T) {
	Convey("UpArticles", t, func() {
		_, _, err := d.UpArticles(context.TODO(), 1, 1, 10)
		So(err, ShouldBeNil)
	})
}
