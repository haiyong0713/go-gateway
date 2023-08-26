package space

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
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
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		flag.Set("conf", "../../cmd/app-interface-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	dao = New(conf.Conf)
	os.Exit(m.Run())
}

// go test -conf="../../app-interface-test.toml"  -v -test.run TestSetting
func TestDao_Setting(t *testing.T) {
	Convey("TestSetting", t, func() {
		setting, err := dao.Setting(context.Background(), 2)
		So(setting, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestDao_Blacklist(t *testing.T) {
	Convey("Blacklist", t, func() {
		gotFs, err := dao.Blacklist(context.Background())
		So(gotFs, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestDao_Report(t *testing.T) {
	Convey("TestReport", t, func() {
		err := dao.Report(context.Background(), 1, "12", "123")
		So(err, ShouldNotBeNil)
	})
}

func TestDao_SpaceMob(t *testing.T) {
	Convey("SpaceMob", t, func() {
		gotMob, err := dao.SpaceMob(context.Background(), 27515257, 27515257, "ios", "")
		Println(gotMob, err)
		So(gotMob, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestDao_ResetTopphoto(t *testing.T) {
	Convey("TestDao_ResetTopphoto", t, func() {
		err := dao.TopphotoReset(context.Background(), "d0a72b4b5b965b82592686616d6381a1", "android", "android")
		Println(err)
		So(err, ShouldBeNil)
	})
}

func TestDao_OfficialDownload(t *testing.T) {
	Convey("TestDao_OfficialDownload", t, func() {
		res, err := dao.OfficialDownload(context.Background(), 1, 1)
		Println(res)
		So(err, ShouldBeNil)
	})
}
