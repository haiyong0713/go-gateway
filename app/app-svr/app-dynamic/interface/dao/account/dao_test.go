package account

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var d *Dao

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-dynamic")
		flag.Set("conf_token", "904b98a0103c506237844db17fb61d45")
		flag.Set("tree_id", "159444")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-dynamic-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func TestDao_IsAttention(t *testing.T) {
	var (
		owners = []int64{27515255}
		uid    = int64(88895133)
	)
	Convey("IsAttention", t, func() {
		res := d.IsAttention(context.TODO(), owners, uid)
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
	})
}
