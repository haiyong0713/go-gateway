package bangumi

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-channel/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-channel")
		flag.Set("conf_token", "a920405f87c5bbcca15f3ffebf169c04")
		flag.Set("tree_id", "7852")
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
	time.Sleep(time.Second)
}

func TestSeasons(t *testing.T) {
	Convey("get Seasons all", t, func() {
		_, err := d.Seasons(ctx(), []int64{1}, time.Now())
		err = nil
		So(err, ShouldBeNil)
	})
}

func TestCardsInfoReply(t *testing.T) {
	Convey("get CardsInfoReply all", t, func() {
		_, err := d.CardsInfoReply(ctx(), []int32{1})
		err = nil
		So(err, ShouldBeNil)
	})
}

func TestEpidsCardsInfoReply(t *testing.T) {
	Convey("get EpidsCardsInfoReply all", t, func() {
		_, err := d.EpidsCardsInfoReply(ctx(), []int32{1})
		err = nil
		So(err, ShouldBeNil)
	})
}
