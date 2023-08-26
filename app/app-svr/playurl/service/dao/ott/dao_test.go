package ott

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/playurl/service/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.playurl-service")
		flag.Set("conf_token", "eec9571409f31d4f8b55a6dfc84d99b8")
		flag.Set("tree_id", "76370")
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
	m.Run()
	os.Exit(0)
}

func TestVideoAuthUgc(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("VideoAuthUgc", t, func(ctx convey.C) {
		can, err := d.VideoAuthUgc(c, 10098982, 10112370)
		ctx.Convey("Then res should be can play.", func(ctx convey.C) {
			fmt.Println(can)
			ctx.So(err, convey.ShouldBeNil)
		})
		cannot, err := d.VideoAuthUgc(c, 10099579, 10113485)
		ctx.Convey("Then res should be cannot play.", func(ctx convey.C) {
			fmt.Println(cannot)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
