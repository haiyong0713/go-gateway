package archive

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	"github.com/smartystreets/goconvey/convey"
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

func TestUpArcs3(t *testing.T) {
	var (
		c   = context.TODO()
		mid = int64(1)
		pn  = 1
		ps  = 5
	)
	convey.Convey("UpArcs3", t, func(ctx convey.C) {
		_, err := d.UpArcs3(c, mid, pn, ps)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestUpCount2(t *testing.T) {
	var (
		c   = context.TODO()
		mid = int64(1)
	)
	convey.Convey("UpCount2", t, func(ctx convey.C) {
		_, err := d.UpCount2(c, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestArchive(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{}
	)
	for i := 0; i < 105; i++ {
		aids = append(aids, 10111164)
	}
	convey.Convey("Archives", t, func(ctx convey.C) {
		arcs, err := d.Archives(c, aids)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
			ctx.Println(arcs)
		})
	})
}

func TestArcs(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{}
	)
	for i := 0; i < 105; i++ {
		aids = append(aids, 10111164)
	}
	convey.Convey("Arcs", t, func(ctx convey.C) {
		arcs, err := d.Arcs(c, aids)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
			ctx.Println(arcs)
		})
	})
}
