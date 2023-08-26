package location

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-wall/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-wall")
		flag.Set("conf_token", "yvxLjLpTFMlbBbc9yWqysKLMigRHaaiJ")
		flag.Set("tree_id", "2283")
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

func TestOrdersUserFlow(t *testing.T) {
	Convey("OrdersUserFlow", t, func() {
		res, err := d.Info(ctx(), "127.0.0.1")
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
