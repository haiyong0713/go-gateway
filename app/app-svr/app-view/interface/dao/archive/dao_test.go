package archive

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"gopkg.in/h2non/gock.v1"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/archive/service/api"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-view")
		flag.Set("conf_token", "3a4CNLBhdFbRQPs7B4QftGvXHtJo92xw")
		flag.Set("tree_id", "4575")
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
func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestArchives(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10113103}
	)
	convey.Convey("Archives", t, func(ctx convey.C) {
		res, err := d.Archives(c, aids, 0, "iphone", "")
		ss, _ := json.Marshal(res)
		fmt.Printf("%s", ss)
		if res[10113103].AttrValV2(api.AttrBitV2ActSeason) == api.AttrYes {
			println(res[10113103].SeasonTheme.BgColor)
			println(res[10113103].SeasonTheme.TextColor)
			println(res[10113103].SeasonTheme.SelectedBgColor)
		}
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestShot(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10114315)
		cid = int64(10165249)
	)
	convey.Convey("Shot", t, func(ctx convey.C) {
		a, _ := d.Shot(c, aid, cid)
		fmt.Printf("%v", a)
	})
}

func TestProgress(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
		mid = int64(1)
	)
	convey.Convey("Progress", t, func(ctx convey.C) {
		_, err := d.Progress(c, aid, mid, "")
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchive(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(-1)
	)
	convey.Convey("Archive", t, func(ctx convey.C) {
		_, err := d.Archive(c, aid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}
