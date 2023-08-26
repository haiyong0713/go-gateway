package ugcpay

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-view/interface/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
)

func init() {
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
	dao = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestAssetRelationDetail(t *testing.T) {
	var (
		c        = context.Background()
		aid      = int64(10111950)
		mid      = int64(2)
		platform = "ios"
	)
	convey.Convey("AssetRelationDetail", t, func(ctx convey.C) {
		asset, err := dao.AssetRelationDetail(c, mid, aid, platform, false)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		ctx.Convey("Then asset should not be nil.", func(ctx convey.C) {
			ctx.So(asset, convey.ShouldNotBeNil)
		})
		fmt.Printf("%+v", asset)
	})
}
