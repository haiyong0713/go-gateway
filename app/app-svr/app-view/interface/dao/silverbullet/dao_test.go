package silverbullet

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model/view"

	"github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
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
	dao = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestRuleCheck(t *testing.T) {
	var (
		c      = context.Background()
		params = &view.SilverEventCtx{
			Action:     "like",
			Aid:        11,
			UpID:       22,
			Mid:        1,
			PubTime:    "2006-01-02 15:04:05",
			LikeSource: "1",
		}
	)
	convey.Convey("RuleCheck", t, func(ctx convey.C) {
		res := dao.RuleCheck(c, params)
		fmt.Println(res)
	})
}
