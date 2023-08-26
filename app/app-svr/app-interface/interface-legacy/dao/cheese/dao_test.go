package cheese

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	"golang.org/x/net/context"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-interface")
		flag.Set("conf_token", "1mWvdEwZHmCYGoXJCVIdszBOPVdtpXb3")
		flag.Set("tree_id", "2688")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-interface-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func TestUserSeason(t *testing.T) {
	convey.Convey("UserSeason", t, func(ctx convey.C) {
		res, total, err := d.UserSeason(context.Background(), 27515255, 1, 2)
		bb, _ := json.Marshal(res)
		fmt.Printf("%s\n", bb)
		fmt.Printf("%d", total)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestEpCards(t *testing.T) {
	convey.Convey("EpCards", t, func(ctx convey.C) {
		res, err := d.EpCards(context.Background(), []int32{15, 16})
		fmt.Printf("%v", res)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
