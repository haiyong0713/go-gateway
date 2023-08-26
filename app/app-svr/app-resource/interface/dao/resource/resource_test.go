package resource

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

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

func TestResSideBar(t *testing.T) {
	Convey("get ResSideBar all", t, func() {
		res, err := d.ResSideBar(ctx())
		ss, _ := json.Marshal(res)
		fmt.Printf("%s", ss)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestEntranceHidden(t *testing.T) {
	Convey("get EntranceHidden all", t, func() {
		res, err := d.EntrancesIsHidden(ctx(), []int64{1, 2}, 9999, 0, "xiaomi")
		ss, _ := json.Marshal(res)
		fmt.Printf("%s", ss)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestSkinConf(t *testing.T) {
	Convey("SkinConf", t, func() {
		res, err := d.SkinConf(ctx())
		Convey("Then err should be nil.", func() {
			So(res, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}

func TestAbTest(t *testing.T) {
	Convey("AbTest", t, func() {
		var groups string
		res, err := d.AbTest(ctx(), groups)
		Convey("Then err should be nil.", func() {
			So(res, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}

func TestEntrancesIsHidden(t *testing.T) {
	Convey("EntrancesIsHidden", t, func() {
		var (
			oids    []int64
			build   int
			plat    int8
			channel string
		)
		res, err := d.EntrancesIsHidden(ctx(), oids, build, plat, channel)
		Convey("Then err should be nil.", func() {
			So(res, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}

func TestPopUp(t *testing.T) {
	Convey("PopUp", t, func() {
		var (
			mid   int64 = 111
			buvid       = ""
			plat  int32 = 1
			build int32 = 111
		)
		res, err := d.PopUp(ctx(), mid, buvid, plat, build)
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
			So(res, ShouldNotBeEmpty)
			Printf("%+v", res)
		})
	})
}
