package share

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model/view"

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

func TestAddShareClick(t *testing.T) {
	var (
		c      = context.Background()
		params = &view.ShareParam{
			AID:          1,
			Build:        8470,
			Platform:     "ios",
			Device:       "phone",
			MobiApp:      "iphone",
			ShareTraceID: "xxx",
		}
		mid   = int64(0)
		buvid = ""
	)
	convey.Convey("AddShareClick", t, func(ctx convey.C) {
		res, err := dao.AddShareClick(c, params, mid, buvid, "")
		fmt.Printf("-----%+v-------", res)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddShareComplete(t *testing.T) {
	var (
		c      = context.Background()
		params = &view.ShareParam{
			AID:          1,
			Build:        8470,
			Platform:     "ios",
			Device:       "phone",
			MobiApp:      "iphone",
			ShareTraceID: "xxx",
		}
		mid   = int64(1)
		buvid = ""
	)
	convey.Convey("AddShareComplete", t, func(ctx convey.C) {
		res, err := dao.AddShareComplete(c, params, mid, buvid)
		fmt.Printf("-----%+v-------", res)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
