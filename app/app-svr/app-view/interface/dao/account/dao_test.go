package account

import (
	"context"
	"flag"
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
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestIsAttention(t *testing.T) {
	convey.Convey("TestIsAttention", t, func(ctx convey.C) {
		isAtten := d.IsAttention(context.TODO(), []int64{9999}, 1684013)
		convey.So(isAtten[9999], convey.ShouldEqual, 0)
	})
}

func TestCard3(t *testing.T) {
	convey.Convey("TestCard3", t, func(ctx convey.C) {
		res, err := d.Card3(context.TODO(), 1684013)
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldNotBeNil)
	})
}

func TestCards3(t *testing.T) {
	convey.Convey("TestCard3", t, func(ctx convey.C) {
		res, err := d.Cards3(context.TODO(), []int64{1684013})
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldNotBeNil)
	})
}

func TestFollowing3(t *testing.T) {
	convey.Convey("TestFollowing3", t, func(ctx convey.C) {
		res, err := d.Following3(context.TODO(), 1684013, 1)
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldNotBeNil)
	})
}
