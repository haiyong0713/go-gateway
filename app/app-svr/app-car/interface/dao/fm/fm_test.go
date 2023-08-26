package fm

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
)

var d *Dao

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-car")
		flag.Set("conf_token", "2c36153a9c62b282e740ae1ba31cd8ad")
		flag.Set("tree_id", "275976")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		dir, _ := filepath.Abs("../../cmd/app-car.toml")
		flag.Set("conf", dir)
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	time.Sleep(time.Second)
}

func TestDao_FmHome(t *testing.T) {
	var (
		c      = context.Background()
		mid    = int64(27515254)
		buvid  = "XY05B900ED5C633B55AC8E94CE6B4D3218938"
		device = model.DeviceInfo{MobiApp: "android_bilithings", Build: 2200001}
		page   = &fm_v2.PageReq{PageSize: 25}
	)
	convey.Convey("FMHome", t, func(ctx convey.C) {
		home, err := d.FmHome(c, mid, buvid, device, page)
		bytes, _ := json.Marshal(home)
		convey.Println(string(bytes))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
