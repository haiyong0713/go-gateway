package tab

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/app-resource/interface/conf"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-resource")
		flag.Set("conf_token", "z8JNX5MFIyDxyBsqwQyF6pnjWQ5YOA14")
		flag.Set("tree_id", "2722")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-resource-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestMenus(t *testing.T) {
	Convey("Menus", t, func() {
		res, err := d.Menus(ctx(), time.Now())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
