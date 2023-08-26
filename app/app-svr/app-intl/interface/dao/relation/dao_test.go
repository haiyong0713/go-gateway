package relation

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-intl/interface/conf"

	"github.com/smartystreets/goconvey/convey"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-intl")
		flag.Set("conf_token", "02007e8d0f77d31baee89acb5ce6d3ac")
		flag.Set("tree_id", "64518")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-intl-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}

func TestStats(t *testing.T) {
	Convey(t.Name(), t, func() {
		var mids = []int64{1, 2, 3}
		res, err := d.Stats(context.Background(), mids)
		Convey("Then isAtten should not be nil.", func() {
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
		})
	})
}

func TestStat(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(1)
	)
	Convey("Stat", t, func() {
		res, err := d.Stat(c, mid)
		Convey("Then isAtten should not be nil.", func() {
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
		})
	})
}

func TestDao_StatsGRPC(t *testing.T) {
	var (
		c    = context.Background()
		mids = []int64{1, 2, 3}
	)
	convey.Convey("StatsGRPC Test", t, func(ctx convey.C) {
		_, err := d.StatsGRPC(c, mids)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
