package archive

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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
	m.Run()
	time.Sleep(time.Second)
}

func ctx() context.Context {
	return context.Background()
}

func TestArchives(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10098214, 10098215, 10098217, 10098218, 10098219, 10098220, 10098221, 10098231, 10098236, 10098239, 10098241, 10098240, 10098232, 10098233, 10098246, 10098243, 10098244, 10098245, 10098247}
	)
	convey.Convey("Archives", t, func(ctx convey.C) {
		res, err := d.Archives(c, aids)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestViews(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10098214, 10098215, 10098217, 10098218, 10098219, 10098220, 10098221, 10098231, 10098236, 10098239, 10098241, 10098240, 10098232, 10098233, 10098246, 10098243, 10098244, 10098245, 10098247}
	)
	convey.Convey("Views", t, func(ctx convey.C) {
		res, err := d.Views(c, aids)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestView(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10098214)
	)
	convey.Convey("View", t, func(ctx convey.C) {
		res, err := d.View(c, aid)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
