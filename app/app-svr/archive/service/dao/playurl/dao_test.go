package playurl

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	batch "git.bilibili.co/bapis/bapis-go/video/vod/playurlugcbatch"
	"go-gateway/app/app-svr/archive/service/api"
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

func TestPlayurlBatch(t *testing.T) {
	var (
		c      = context.TODO()
		cidArr = []*batch.RequestVideoItem{}
	)
	cidArr = append(cidArr, &batch.RequestVideoItem{Cid: 10108828, IsSp: false})
	convey.Convey("PlayurlBatch", t, func(ctx convey.C) {
		res, err := d.PlayurlBatch(c, cidArr, nil, 0, true)
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestPlayurlVolume(t *testing.T) {
	var (
		c        = context.TODO()
		cids     = []uint64{10341397, 10341396}
		batchArg = &api.BatchPlayArg{
			Mid: 123,
			Ip:  "123",
		}
	)
	convey.Convey("TestPlayurlVolume", t, func(ctx convey.C) {
		res, err := d.PlayurlVolume(c, cids, batchArg)
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
