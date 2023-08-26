package archive

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-intl/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-intl")
		flag.Set("conf_token", "02007e8d0f77d31baee89acb5ce6d3ac")
		flag.Set("tree_id", "64518")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-intl-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestArchives(t *testing.T) {
	Convey(t.Name(), t, func() {
		_, err := d.Archives(context.Background(), []int64{1, 2})
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestArchive(t *testing.T) {
	Convey(t.Name(), t, func() {
		_, err := d.Archive(context.Background(), 122)
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestProgress(t *testing.T) {
	Convey(t.Name(), t, func() {
		_, err := d.Progress(context.Background(), 12, 133)
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}
