package channel

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-view/interface/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-view")
		flag.Set("conf_token", "3a4CNLBhdFbRQPs7B4QftGvXHtJo92xw")
		flag.Set("tree_id", "4575")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-view-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestResourceChannels(t *testing.T) {
	var (
		c = context.TODO()
	)
	convey.Convey("Info", t, func(ctx convey.C) {
		_, err := d.ResourceChannels(c, 10200266, 27515250, 3)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestChannelHonor(t *testing.T) {
	var (
		c = context.TODO()
	)
	convey.Convey("Info", t, func(ctx convey.C) {
		res, err := d.ChannelHonor(c, 280009482)
		fmt.Printf("%+v", res)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
