package duertv

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-car/job/conf"
	"go-gateway/app/app-svr/app-car/job/model/duertv"

	"github.com/glycerine/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-car-job")
		flag.Set("conf_token", "25bd6013b6c2911ad5eda592affa0b86")
		flag.Set("tree_id", "291964")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		dir, _ := filepath.Abs("../../cmd/app-car-job.toml")
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

func TestPush(t *testing.T) {
	var (
		c    = context.TODO()
		data []*duertv.DuertvPush
		now  time.Time
	)
	convey.Convey("Push", t, func(ctx convey.C) {
		err := d.Push(c, data, now)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
