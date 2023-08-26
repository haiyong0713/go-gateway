package resource

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-common/library/database/sql"
	"go-gateway/app/app-svr/resource/service/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.resource-service")
		flag.Set("conf_token", "a1bf4b2063965fbc2345edb9ab11baf8")
		flag.Set("tree_id", "3232")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/resource-service-test.toml")
	}
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("resource-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	d = New(cfg)
	m.Run()
	os.Exit(0)
}

func WithReopenDB(f func(d *Dao)) func() {
	return func() {
		convey.Reset(func() {
			d.db = sql.NewMySQL(d.c.DB.Res)
		})
		f(d)
	}
}

func TestDaoPing(t *testing.T) {
	convey.Convey("Ping", t, func(ctx convey.C) {
		err := d.Ping(context.TODO())
		ctx.Convey("Err should be nil", func() {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
