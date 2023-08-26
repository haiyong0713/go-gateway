package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-common/component/metadata/device"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
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
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestVideoGuide(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(123)
		cid = int64(123)
		mid = int64(1)
	)
	convey.Convey("VideoGuide", t, func(ctx convey.C) {
		res, err := d.VideoGuide(c, aid, cid, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			fmt.Printf("%+v", res)
			err = nil
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestClickPlayerCard(t *testing.T) {
	var (
		c   = context.TODO()
		id  = int64(123)
		aid = int64(123)
		cid = int64(123)
		mid = int64(123)
		arg = &viewApi.ClickPlayerCardReq{
			Aid: aid,
			Cid: cid,
			Id:  id,
		}
		dev = device.Device{
			RawPlatform: "ios",
			RawMobiApp:  "iphone",
			Device:      "phone",
			Build:       999999,
		}
	)
	convey.Convey("ClickPlayerCard", t, func(ctx convey.C) {
		err := d.ClickPlayerCard(c, arg, mid, dev)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			fmt.Println(err)
			err = nil
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
