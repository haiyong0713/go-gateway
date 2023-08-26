package videoup_open

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-interface")
		flag.Set("conf_token", "1mWvdEwZHmCYGoXJCVIdszBOPVdtpXb3")
		flag.Set("tree_id", "2688")
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

// TestAndroidCreative dao ut.
func TestAndroidCreative(t *testing.T) {
	var (
		c   = context.TODO()
		mid = int64(2)
	)
	convey.Convey("AndroidCreative", t, func(ctx convey.C) {
		res, err := d.AndroidCreative(c, mid, 5390300)
		r, _ := json.Marshal(res)
		fmt.Printf("%s", r)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

// IOSCreative dao ut.
func TestIOSCreative(t *testing.T) {
	var (
		c   = context.TODO()
		mid = int64(1)
	)
	convey.Convey("IOSCreative", t, func(ctx convey.C) {
		res, err := d.IOSCreative(c, mid, 9999)
		r, _ := json.Marshal(res)
		fmt.Printf("%s", r)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_Creative(t *testing.T) {
	var (
		c   = context.TODO()
		mid = int64(1)
	)
	convey.Convey("TestDao_Creative", t, func(ctx convey.C) {
		isUp, show, err := d.Creative(c, mid)
		fmt.Printf("%d %d", isUp, show)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
