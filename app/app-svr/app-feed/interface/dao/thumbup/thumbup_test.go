package thumbup

import (
	"context"
	"flag"
	"os"
	"testing"

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
	} else {
		flag.Set("conf", "../../cmd/app-feed-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
}

func TestHasLike(t *testing.T) {
	convey.Convey("Ping", t, func(ctx convey.C) {
		var (
			mid        int64
			messageIDs []int64
			c          = context.TODO()
		)
		convey.Convey("Ping", t, func(ctx convey.C) {
			_, err := d.HasLike(c, mid, messageIDs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func ctx() context.Context {
	return context.Background()
}
