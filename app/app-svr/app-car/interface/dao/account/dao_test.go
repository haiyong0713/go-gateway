package account

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-car/interface/conf"

	"github.com/glycerine/goconvey/convey"
)

var (
	d *Dao
)

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
	time.Sleep(time.Second)
}

func ctx() context.Context {
	return context.Background()
}

func TestProfile3(t *testing.T) {
	convey.Convey("Profile3", t, func(ctx convey.C) {
		var (
			c   = context.TODO()
			mid = int64(12321312)
		)
		_, err := d.Profile3(c, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
