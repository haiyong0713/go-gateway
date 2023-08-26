package elec

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-view/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-show")
		flag.Set("conf_token", "Pae4IDOeht4cHXCdOkay7sKeQwHxKOLA")
		flag.Set("tree_id", "2687")
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
	time.Sleep(time.Second)
}

func ctx() context.Context {
	return context.Background()
}

func TestInfo(t *testing.T) {
	Convey("get Info all", t, func() {
		res, err := d.Info(ctx(), 1, 1)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestRankElecMonthUP(t *testing.T) {
	Convey("get Info all", t, func() {
		res, err := d.RankElecMonthUPList(ctx(), []int64{111005048}, 601000, "iphone", "phone", "")
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
		fmt.Printf("%v", res)
	})
}
