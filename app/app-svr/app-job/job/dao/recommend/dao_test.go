package recommend

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-job/job/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-job")
		flag.Set("conf_token", "613aae0ddd1cc47a79920d6115cea472")
		flag.Set("tree_id", "2861")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-job-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func ctx() context.Context {
	return context.Background()
}

func TestRecommend(t *testing.T) {
	convey.Convey("Recommend", t, func() {
		res, err := d.Recommend(ctx())
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldNotBeEmpty)
	})
}
