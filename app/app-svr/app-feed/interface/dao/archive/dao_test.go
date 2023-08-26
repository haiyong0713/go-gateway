package archive

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestPing(t *testing.T) {
	var (
		c = context.TODO()
	)
	convey.Convey("Ping", t, func(ctx convey.C) {
		d.PingMC(c)
	})
}

func TestArchives(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10098214, 10098215, 10098217, 10098218, 10098219, 10098220, 10098221, 10098231, 10098236, 10098239, 10098241, 10098240, 10098232, 10098233, 10098246, 10098243, 10098244, 10098245, 10098247}
	)
	convey.Convey("Ping", t, func(ctx convey.C) {
		res, err := d.Archives(c, aids, 0, "", "")
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
