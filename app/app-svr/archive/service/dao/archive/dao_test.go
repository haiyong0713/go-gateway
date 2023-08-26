package archive

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/net/trace"

	"go-gateway/app/app-svr/archive/service/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app", "archive.service")
		flag.Set("appid", "main.app-svr.archive-service")
		flag.Set("conf_token", "Y2LJhIsHx87nJaOBSxuG5TeZoLdBFlrE")
		flag.Set("tree_id", "2302")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	trace.Init(nil)
	defer trace.Close()
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestArc(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(-1)
	)
	convey.Convey("Arc", t, func(ctx convey.C) {
		_, err := d.Arc(c, aid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveVideos3(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10097272)
	)
	convey.Convey("Videos3", t, func(ctx convey.C) {
		vs, err := d.Videos3(c, aid)
		ctx.Convey("Then err should be nil.vs should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(vs, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveVideosByAids3(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10097272, 10098500}
	)
	convey.Convey("VideosByAids3", t, func(ctx convey.C) {
		vs, err := d.VideosByAids3(c, aids)
		ctx.Convey("Then err should be nil.vs should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(vs, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveVideo3(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(320006469)
		cid = int64(10342767)
	)
	convey.Convey("Video3", t, func(ctx convey.C) {
		v, err := d.Video3(c, aid, cid)
		ctx.Convey("Then err should not be nil.v should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
			ctx.So(v, convey.ShouldBeNil)
		})
	})
}

func TestArchiveUpVideo3(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
		cid = int64(2)
	)
	convey.Convey("UpVideo3", t, func(ctx convey.C) {
		v, err := d.Video3(c, aid, cid)
		ctx.Convey("Then err should not be nil.v should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
			ctx.So(v, convey.ShouldBeNil)
		})
	})
}
