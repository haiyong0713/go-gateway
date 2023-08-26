package tag

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/web-svr/activity/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
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
		flag.Set("conf", "../../cmd/activity-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}

func TestTagByName(t *testing.T) {
	Convey("RawBases", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			rly, err := d.TagByName(c, "uptwo")
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				fmt.Printf("%v", rly)
			})
		})
	})
}

// AddTag
func TestAddTag(t *testing.T) {
	Convey("RawBases", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			rly, err := d.AddTag(c, "upTwo", 1)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				fmt.Printf("%v", rly)
			})
		})
	})
}
