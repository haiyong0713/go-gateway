package popular

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go-common/library/conf/paladin.v2"
	"go-gateway/app/web-svr/native-page/interface/conf"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.web-svr.activity")
		flag.Set("conf_token", "22edc93e2998bf0cb0bbee661b03d41f")
		flag.Set("tree_id", "2873")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/native-page-interface-test.toml")
	}
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	cfg := &conf.Config{}
	if err := paladin.Get("native-page-interface.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	d = New(cfg)
	os.Exit(m.Run())
}

func TestTimeLine(t *testing.T) {
	convey.Convey("preListKey", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			rly, p1 := d.PageArcs(context.Background(), 0, 2, 0)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				str, _ := json.Marshal(rly)
				fmt.Printf("%v", string(str))
				fmt.Printf("error %v", p1)
			})
		})
	})
}
