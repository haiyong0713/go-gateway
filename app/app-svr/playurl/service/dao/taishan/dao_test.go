package taishan

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-common/library/net/trace"

	"go-gateway/app/app-svr/playurl/service/conf"
	tmdl "go-gateway/app/app-svr/playurl/service/model/taishan"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.playurl-service")
		flag.Set("conf_token", "eec9571409f31d4f8b55a6dfc84d99b8")
		flag.Set("tree_id", "76370")
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
	// taishan缓存必须初始化trace，如果kv拿不到trace将无法校验通过
	trace.Init(conf.Conf.Tracer)
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestPlayConfGet(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("PlayConfGet", t, func(ctx convey.C) {
		rly, err := d.PlayConfGet(c, "45a31014725e1987_yyyy")
		ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			if rly != nil {
				fmt.Printf("PlayConfGet %v", rly)
			}
		})
	})
}

// PlayConfSet
func TestPlayConfSet(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("PlayConfGet", t, func(ctx convey.C) {
		err := d.PlayConfSet(c, &tmdl.PlayConfs{PlayConfs: map[int64]*tmdl.PlayConf{1: {Show: true}}}, "45a31014725e1987")
		ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
