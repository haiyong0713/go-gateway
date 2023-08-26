package dao

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin"
	appmodel "go-gateway/app/app-svr/misaka/interface/internal/model/app"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-misaka")
		flag.Set("conf_token", "213ceacc360006f572dd0daeb4887b0d")
		flag.Set("tree_id", "90031")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../configs")
	}
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	ac := new(paladin.TOML)
	if err := paladin.Watch("application.toml", ac); err != nil {
		panic(err)
	}
	d = New(ac)
	os.Exit(m.Run())
}

func TestPub(t *testing.T) {
	convey.Convey("pub", t, func(ctx convey.C) {
		err := d.PubApp(context.Background(), &appmodel.Info{
			Data: &appmodel.Data{
				LogID:    1,
				MobiApp:  "android",
				Device:   "android",
				Platform: "android",
				Buvid:    "woshibuvid",
				Brand:    "haha",
				Model:    "hahamodel",
				Osver:    "10101010",
				Build:    "12345",
				Network:  1,
				Mid:      123456,
			},
			IP: "0.0.0.0",
		})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestClose(t *testing.T) {
	convey.Convey("Close", t, func(ctx convey.C) {
		ctx.Convey("Close", func(ctx convey.C) {
			d.Close()
		})
	})
}

func TestPing(t *testing.T) {
	convey.Convey("Ping", t, func(ctx convey.C) {
		ctx.Convey("Ping", func(ctx convey.C) {
			err := d.Ping(context.Background())
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
