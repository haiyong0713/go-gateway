package ad

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
	d *Dao
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
	} else {
		flag.Set("conf", "../../cmd/app-interface-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func TestDao_AdVTwo(t *testing.T) {
	Convey("GetUserJumpBBQGrant", t, func() {
		gotFs, err := d.AdVTwo(context.Background(), 111004263, 111004263, 8910, "", []int64{}, "", "", "", "", "", "")
		fmt.Print(gotFs)
		So(err, ShouldBeNil)
	})
}

func TestDao_CreatedTopicList(t *testing.T) {
	var (
		mid = int64(12154415)
	)
	Convey("CreatedTopicList", t, func() {
		gotFs, err := d.CreatedTopicList(context.Background(), mid)
		bs, _ := json.Marshal(gotFs)
		fmt.Print(string(bs))
		So(err, ShouldBeNil)
	})
}
