package season

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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

func TestSeason(t *testing.T) {
	var (
		c        = context.Background()
		seasonID = int64(1099)
	)
	convey.Convey("Season", t, func(ctx convey.C) {
		season, err := d.Season(c, seasonID)
		ctx.Convey("Then err should be nil.season should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(season, convey.ShouldNotBeNil)
		})
		qq, _ := json.Marshal(season)
		fmt.Println(string(qq))
	})
}
