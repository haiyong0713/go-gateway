package dynamic

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

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
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

func TestDynVideoList(t *testing.T) {
	var (
		c                                = context.TODO()
		uid                              int64
		updateBaseLine, assistBaseLine   string
		dynType                          []string
		attention                        *dyncommongrpc.AttentionInfo
		build                            int
		platform, mobiApp, buvid, device string
	)
	convey.Convey("DynVideoList", t, func(ctx convey.C) {
		res, err := d.DynVideoList(c, uid, updateBaseLine, assistBaseLine, dynType, attention, build, platform, mobiApp, buvid, device)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDynVideoHistory(t *testing.T) {
	var (
		c                                = context.TODO()
		uid                              int64
		offset                           string
		page                             int64
		dynType                          []string
		attention                        *dyncommongrpc.AttentionInfo
		build                            int
		platform, mobiApp, buvid, device string
	)
	convey.Convey("DynVideoHistory", t, func(ctx convey.C) {
		res, err := d.DynVideoHistory(c, uid, offset, page, dynType, attention, build, platform, mobiApp, buvid, device)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
