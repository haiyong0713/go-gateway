package bangumi

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-car/job/conf"

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

func TestModule(t *testing.T) {
	var (
		c          = context.TODO()
		pn         = 1
		ps         = 20
		seasonType = 1
		bsource    = "xiaodu"
	)
	convey.Convey("ChannelContent", t, func(ctx convey.C) {
		res, err := d.ChannelContent(c, pn, ps, seasonType, bsource)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestChannelContentChange(t *testing.T) {
	var (
		c          = context.TODO()
		pn         = 1
		ps         = 20
		seasonType = 1
		bsource    = "xiaodu"
	)
	convey.Convey("ChannelContentChange", t, func(ctx convey.C) {
		now := time.Now()
		thisDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		// 前一天
		yesterdayStart := thisDay.AddDate(0, 0, -1)
		// 前一天的23:59:59
		yesterdayEnd := time.Date(yesterdayStart.Year(), yesterdayStart.Month(), yesterdayStart.Day(), 23, 59, 59, 0, now.Location())
		res, err := d.ChannelContentChange(c, pn, ps, seasonType, bsource, yesterdayStart, yesterdayEnd)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
