package account

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
)

// TestMain dao ut.
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
	dao = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func ctx() context.Context {
	return context.Background()
}

// TestDaoBlockTime dao ut.
func TestDaoBlockTime(t *testing.T) {
	Convey("get account", t, func() {
		_, err := dao.BlockTime(ctx(), 2089809)
		So(err, ShouldBeNil)
	})
}

// TestDaoProfile3 dao ut.
func TestDaoProfile3(t *testing.T) {
	Convey("get account", t, func() {
		res, err := dao.Profile3(ctx(), 1111112587)
		fmt.Println(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestCard dao ut.
func TestCard(t *testing.T) {
	Convey("get account", t, func() {
		res, err := dao.Card(ctx(), 2089809)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestProfileByName3 dao ut.
func TestProfileByName3(t *testing.T) {
	Convey("get account", t, func() {
		_, err := dao.ProfileByName3(ctx(), "冠冠爱看书")
		So(err, ShouldNotBeNil)
		//So(res, ShouldNotBeEmpty)
	})
}

// TestInfos3 dao ut.
func TestInfos3(t *testing.T) {
	Convey("get account", t, func() {
		res, err := dao.Infos3(ctx(), []int64{2089809})
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestRelations3 dao ut.
func TestRelations3(t *testing.T) {
	Convey("get account", t, func() {
		res := dao.Relations3(ctx(), []int64{111005047, 111005047, 111005047, 111005043}, 15555180)
		ss, _ := json.Marshal(res)
		fmt.Printf("%s", ss)
		So(res, ShouldNotBeEmpty)
	})
}

// TestRichRelations3 dao ut.
func TestRichRelations3(t *testing.T) {
	Convey("get account", t, func() {
		res, err := dao.RichRelations3(ctx(), 111005046, 15555180)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestCards3 dao ut.
func TestCards3(t *testing.T) {
	Convey("get account", t, func() {
		res, err := dao.Cards3(ctx(), []int64{2089809})
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestPrompting dao ut.
func TestPrompting(t *testing.T) {
	Convey("get Prompting", t, func() {
		res, err := dao.Prompting(ctx(), 15555180)
		fmt.Printf("-----%+v-----", res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}
