package history

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var d *Dao

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

func TestDao_ArchiveInfo(t *testing.T) {
	Convey("ArchiveInfo", t, func() {
		_, err := d.Archive(context.TODO(), []int64{1, 2})
		So(err, ShouldBeNil)
	})
}

func TestDao_GetList(t *testing.T) {
	Convey("GetList", t, func() {
		_, err := d.History(context.TODO(), 27515256, 1, 20)
		So(err, ShouldBeNil)
	})
}

func TestDao_Cursor(t *testing.T) {
	Convey("Cursor", t, func() {
		res, err := d.Cursor(context.TODO(), 111004852, 0, 20, "", []string{"archive", "pgc", "article", "live", "article-list", "cheese"}, "")
		fmt.Printf("%+v", res)
		So(err, ShouldBeNil)
	})
}

func TestDao_Search(t *testing.T) {
	Convey("Search", t, func() {
		res, _, err := d.Search(context.TODO(), 111004852, 1, 20, " ", []string{"archive", "pgc", "article", "live", "article-list", "cheese"})
		fmt.Printf("%+v", res)
		So(err, ShouldBeNil)
	})
}
