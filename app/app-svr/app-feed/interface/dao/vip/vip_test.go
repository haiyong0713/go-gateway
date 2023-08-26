package vip

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	viprpc "git.bilibili.co/bapis/bapis-go/vip/service"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-feed")
		flag.Set("conf_token", "OC30xxkAOyaH9fI6FRuXA0Ob5HL0f3kc")
		flag.Set("tree_id", "2686")
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

func TestTipsRenew(t *testing.T) {
	Convey("get TipsRenew all", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			platform = int64(1)
			build    = int(5405000)
			mid      = int64(14135892)
			res      *viprpc.TipsRenewReply
			err      error
		)
		defer mockCtrl.Finish()
		mockArc := viprpc.NewMockVipClient(mockCtrl)
		d.rpcClient = mockArc
		mockArc.EXPECT().TipsRenew(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.TipsRenew(ctx(), build, platform, mid)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
