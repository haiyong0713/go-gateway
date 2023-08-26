package resource

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	"github.com/smartystreets/goconvey/convey"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-feed")
		flag.Set("conf_token", "OC30xxkAOyaH9fI6FRuXA0Ob5HL0f3kc")
		flag.Set("tree_id", "2686")
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

func TestBanner(t *testing.T) {
	Convey("Banner", t, func(cc convey.C) {
		var (
			plat                                             int8
			build                                            int
			mid                                              int64
			resIDs, channel, buvid, network, mobiApp, device string
			isAd                                             bool
			openEvent, adExtra, hash                         string
		)
		_, _, err := d.Banner(ctx(), plat, build, mid, resIDs, channel, buvid, network, mobiApp, device, isAd, openEvent, adExtra, hash)
		cc.Convey("Then err should be nil.list should not be nil.", func(cc convey.C) {
			cc.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAbTest(t *testing.T) {
	Convey("AbTest", t, func(cc convey.C) {
		var (
			groups string
		)
		_, err := d.AbTest(ctx(), groups)
		cc.Convey("Then err should be nil.list should not be nil.", func(cc convey.C) {
			cc.So(err, convey.ShouldBeNil)
		})
	})
}
