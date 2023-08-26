package channel

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	"github.com/smartystreets/goconvey/convey"
	. "github.com/smartystreets/goconvey/convey"

	"testing"
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
	if os.Getenv("UT_LOCAL_TEST") != "" {
		dir, _ := filepath.Abs("../../cmd/app-feed-test.toml")
		flag.Set("conf", dir)
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestChannels(t *testing.T) {
	Convey("Channels", t, func(cc convey.C) {
		var (
			channelIDs []int64
			mid        int64
		)
		_, err := d.Channels(ctx(), channelIDs, mid)
		cc.Convey("Then err should be nil.", func(cc convey.C) {
			cc.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDetails(t *testing.T) {
	Convey("Details", t, func(cc convey.C) {
		var (
			tids []int64
		)
		_, err := d.Details(ctx(), tids)
		cc.Convey("Then err should be nil.", func(cc convey.C) {
			cc.So(err, convey.ShouldBeNil)
		})
	})
}
