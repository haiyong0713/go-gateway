package history

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/archive/service/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.archive-service")
		flag.Set("conf_token", "Y2LJhIsHx87nJaOBSxuG5TeZoLdBFlrE")
		flag.Set("tree_id", "2302")
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

func TestProgress(t *testing.T) {
	var (
		c     = context.TODO()
		aids  = []int64{720002823}
		mid   = int64(27515256)
		buvid = ""
	)
	convey.Convey("Progress", t, func(ctx convey.C) {
		res, err := d.Progress(c, aids, mid, buvid)
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
