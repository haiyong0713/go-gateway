package dynamic

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/archive-honor/service/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.archive-honor-service")
		flag.Set("conf_token", "6a91870821701a2c4e6b49d7fc270af2")
		flag.Set("tree_id", "136937")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/archive-honor-service.toml")
	}
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	cfg := &conf.Config{}
	if err := paladin.Get("archive-honor-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	d = New(cfg)
	m.Run()
	os.Exit(0)
}

func TestGetMsgKey(t *testing.T) {
	var (
		c        = context.TODO()
		arcTitle = "我是稿件名"
		upName   = "我是up名"
		//senderUID = 88895034
		//NotifyCode = "3_8"
	)
	convey.Convey("GetMsgKey", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			msgKey, err := d.GetMsgKey(c, arcTitle, upName)
			fmt.Printf("%d", msgKey)
			ctx.So(msgKey, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSendMsgURL(t *testing.T) {
	var (
		c      = context.TODO()
		upid   = uint64(15555180)
		msgKey = uint64(6769531888883204097)
	)
	convey.Convey("SendMsgURL", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			err := d.SendMsg(c, upid, msgKey)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
