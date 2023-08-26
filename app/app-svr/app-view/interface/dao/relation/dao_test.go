package relation

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-view/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-show")
		flag.Set("conf_token", "Pae4IDOeht4cHXCdOkay7sKeQwHxKOLA")
		flag.Set("tree_id", "2687")
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

func TestPrompt(t *testing.T) {
	Convey("Prompt", t, func() {
		_, err := d.Prompt(context.TODO(), 1, 1, 1)
		So(err, ShouldBeNil)
	})
}

func TestStat(t *testing.T) {
	Convey("Stat", t, func() {
		_, err := d.Stat(context.TODO(), 1)
		So(err, ShouldBeNil)
	})
}

func TestStatsGRPC(t *testing.T) {
	Convey("StatsGRPC", t, func() {
		_, err := d.StatsGRPC(context.TODO(), []int64{1})
		So(err, ShouldBeNil)
	})
}

// Relation
func TestRelation(t *testing.T) {
	Convey("Relation", t, func() {
		_, err := d.Relation(context.TODO(), 15555180, 27515412)
		So(err, ShouldBeNil)
	})
}
