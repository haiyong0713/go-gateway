package player

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-player/interface/conf"
	"go-gateway/app/app-svr/app-player/interface/model"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-player")
		flag.Set("conf_token", "e477d98a7c5689623eca4f32f6af735c")
		flag.Set("tree_id", "52581")
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

func TestPlayURLV2(t *testing.T) {
	var (
		c      = context.Background()
		params = &model.Param{
			AID:   10112294,
			CID:   10154057,
			Qn:    120,
			Fourk: 1,
		}
		mid = int64(1)
	)
	convey.Convey("PlayURLV2", t, func(ctx convey.C) {
		arc, err := d.PlayURLV2(c, params, mid, 0, 0)
		fmt.Printf("%+v", arc)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
