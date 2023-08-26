package redis

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/player-online/internal/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
)

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.player-online")
		flag.Set("conf_token", "1976c6f5fc597b9706727c14b3c70518")
		flag.Set("tree_id", "588149")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../../cmd/player-online-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	dao = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestSetOnlineCountCache(t *testing.T) {
	var (
		c           = context.Background()
		aid   int64 = 560008223
		cid   int64 = 10300412
		exp   int64 = 180
		value int64 = 100
	)
	convey.Convey("SetOnlineCountCache", t, func(ctx convey.C) {
		err := dao.SetOnlineCountCache(c, aid, cid, exp, value)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
