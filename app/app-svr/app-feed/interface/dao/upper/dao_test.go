package upper

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

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
	time.Sleep(time.Second)
}

func Test_Feed(t *testing.T) {
	convey.Convey("Feed", t, func(ctx convey.C) {
		var (
			c      = context.TODO()
			mid    int64
			pn, ps int
		)
		_, err := d.Feed(c, mid, pn, ps)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_ArchiveFeed(t *testing.T) {
	convey.Convey("ArchiveFeed", t, func(ctx convey.C) {
		var (
			c      = context.TODO()
			mid    int64
			pn, ps int
		)
		_, err := d.ArchiveFeed(c, mid, pn, ps)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_AppUnreadCount(t *testing.T) {
	convey.Convey("AppUnreadCount", t, func(ctx convey.C) {
		var (
			c              = context.TODO()
			mid            int64
			withoutBangumi bool
		)
		_, err := d.AppUnreadCount(c, mid, withoutBangumi)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_ArticleFeed(t *testing.T) {
	convey.Convey("ArticleFeed", t, func(ctx convey.C) {
		var (
			c      = context.TODO()
			mid    int64
			pn, ps int
		)
		_, err := d.ArticleFeed(c, mid, pn, ps)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_ArticleUnreadCount(t *testing.T) {
	convey.Convey("ArticleUnreadCount", t, func(ctx convey.C) {
		var (
			c   = context.TODO()
			mid int64
		)
		_, err := d.ArticleUnreadCount(c, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
