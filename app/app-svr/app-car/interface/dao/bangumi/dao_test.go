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

func TestModule(t *testing.T) {
	var (
		c       = context.TODO()
		id      = 207
		mobiApp = "android_bilithings"
	)
	convey.Convey("Module", t, func(ctx convey.C) {
		res, err := d.Module(c, id, mobiApp, "")
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestView(t *testing.T) {
	var (
		c         = context.TODO()
		mid       = int64(1111)
		seasonid  = int64(34924)
		mobiApp   = "android_bilithings"
		accessKey = "ea912577f2a79b6286b720fb7adb4c81"
		platform  = "android"
		buvid     = "test123"
		build     = 123
	)
	convey.Convey("View", t, func(ctx convey.C) {
		res, err := d.View(c, mid, seasonid, accessKey, "", mobiApp, platform, buvid, "", build)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCards(t *testing.T) {
	var (
		c         = context.TODO()
		seasonids = []int32{34924}
	)
	convey.Convey("Cards", t, func(ctx convey.C) {
		res, err := d.Cards(c, seasonids)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestEpCards(t *testing.T) {
	var (
		c     = context.TODO()
		epids = []int32{34924}
	)
	convey.Convey("EpCards", t, func(ctx convey.C) {
		res, err := d.EpCards(c, epids)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestMyRelations(t *testing.T) {
	var (
		c   = context.TODO()
		mid = int64(3123)
	)
	convey.Convey("MyRelations", t, func(ctx convey.C) {
		res, err := d.MyRelations(c, mid)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestMyFollows(t *testing.T) {
	var (
		c          = context.TODO()
		mid        = int64(3123)
		followType = 1
		pn         = 1
		ps         = 20
	)
	convey.Convey("MyFollows", t, func(ctx convey.C) {
		res, err := d.MyFollows(c, mid, followType, pn, ps)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
