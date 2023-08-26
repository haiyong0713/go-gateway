package resource

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-intl/interface/conf"
	"go-gateway/app/app-svr/app-intl/interface/model"

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

func TestPlayerIcon(t *testing.T) {
	Convey(t.Name(), t, func() {
		res, err := d.PlayerIcon(context.Background(), 0, []int64{}, 0)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestPasterCID(t *testing.T) {
	Convey(t.Name(), t, func() {
		res, err := d.PasterCID(context.Background())
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestBanner(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			plat          = model.PlatIPhone
			build         = 9999
			mid           = int64(1)
			resIDs        = "467"
			channel       = "ios"
			buvid         = ""
			network       = "wifi"
			mobiApp       = "iphone"
			device        = "phone"
			isAd          = false
			openEvent     = "cold"
			adExtra, hash string
		)
		res, _, err := d.Banner(context.Background(), plat, build, mid, resIDs, channel, buvid, network, mobiApp, device, isAd, openEvent, adExtra, hash)
		Convey("Then err should be nil", func() {
			So(res, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}
