package location

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app", "feed-admin")
		flag.Set("app_id", "main.web-svr.feed-admin")
		flag.Set("conf_token", "e0d2b216a460c8f8492473a2e3cdd218")
		flag.Set("tree_id", "45266")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		flag.Set("conf", "../../cmd/feed-admin-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}

func TestAddPolicy(t *testing.T) {
	convey.Convey("AddPolicy", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			pid, err := d.AddPolicy(context.Background(), []int64{2831, 2833})
			fmt.Println(pid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPolicyInfo(t *testing.T) {
	convey.Convey("PolicyInfo", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			areaIDs, err := d.PolicyInfo(context.Background(), 4363)
			fmt.Println(areaIDs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPolicyInfos(t *testing.T) {
	convey.Convey("PolicyInfos", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, err := d.PolicyInfos(context.Background(), []int64{4363, 4364})
			ress, _ := json.Marshal(res)
			fmt.Printf("%s", ress)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
