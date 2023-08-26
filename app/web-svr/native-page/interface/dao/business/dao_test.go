package business

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/web-svr/native-page/interface/conf"

	"gopkg.in/h2non/gock.v1"

	"github.com/smartystreets/goconvey/convey"
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
	d.client.SetTransport(gock.DefaultTransport)
	os.Exit(m.Run())
}

func TestProductDetail(t *testing.T) {
	convey.Convey("ProductDetail", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reply, err := d.ProductDetail(c, "427468146283003900", 0, 15)
			convCtx.Convey("Then err should be nil.reply should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", reply)
			})
		})
	})
}

func TestSourceDetail(t *testing.T) {
	convey.Convey("SourceDetail", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reply, err := d.SourceDetail(c, "427468146283003900", 0, 15)
			convCtx.Convey("Then err should be nil.reply should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", reply)
			})
		})
	})
}
