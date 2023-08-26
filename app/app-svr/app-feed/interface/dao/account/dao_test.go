package account

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-feed")
		flag.Set("conf_token", "OC30xxkAOyaH9fI6FRuXA0Ob5HL0f3kc")
		flag.Set("tree_id", "2686")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		dir, _ := filepath.Abs("../../cmd/app-feed-test.toml")
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

func TestCards3GRPC(t *testing.T) {
	convey.Convey("Cards3GRPC", t, func(ctx convey.C) {
		var (
			c    = context.TODO()
			mids = []int64{12321312}
		)
		_, err := d.Cards3GRPC(c, mids)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRelations3GRPC(t *testing.T) {
	convey.Convey("Relations3GRPC", t, func(ctx convey.C) {
		var (
			c      = context.TODO()
			owners = []int64{12321312}
			mid    = int64(12321312)
		)
		res := d.Relations3GRPC(c, owners, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestIsAttention(t *testing.T) {
	convey.Convey("IsAttentionGRPC", t, func(ctx convey.C) {
		var (
			c      = context.TODO()
			owners = []int64{12321312}
			mid    = int64(12321312)
		)
		res := d.IsAttentionGRPC(c, owners, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
