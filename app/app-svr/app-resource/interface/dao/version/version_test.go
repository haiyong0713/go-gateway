package version

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

func TestAll(t *testing.T) {
	Convey("get all", t, func() {
		res, err := d.All(ctx())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestUpdates(t *testing.T) {
	Convey("get Updates all", t, func() {
		res, err := d.Updates(ctx())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestIncrementals(t *testing.T) {
	Convey("get Incrementals all", t, func() {
		res, err := d.Incrementals(ctx())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestSos(t *testing.T) {
	Convey("get Sos all", t, func() {
		res, err := d.Sos(ctx())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestRn(t *testing.T) {
	Convey("get Rn all", t, func() {
		res, err := d.Rn(ctx())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestLimits(t *testing.T) {
	Convey("Limits", t, func() {
		res, err := d.Limits(ctx())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func ctx() context.Context {
	return context.Background()
}
