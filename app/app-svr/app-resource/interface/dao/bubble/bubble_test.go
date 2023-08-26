package bubble

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
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
	os.Exit(m.Run())
}

func TestBubble(t *testing.T) {
	Convey("Bubble", t, func() {
		res, err := d.Bubble(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestSetBubbleConfig(t *testing.T) {
	Convey("SetBubbleConfig", t, func() {
		var (
			buid, mid int64
			state     int
			expire    int32
		)
		err := d.SetBubbleConfig(ctx(), buid, mid, state, expire)
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestBubbleConfig(t *testing.T) {
	Convey("BubbleConfig", t, func() {
		var (
			buid, mid int64
		)
		_, err := d.BubbleConfig(ctx(), buid, mid)
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}
