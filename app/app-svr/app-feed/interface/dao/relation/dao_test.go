package relation

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"

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

func TestStatsGRPC(t *testing.T) {
	Convey("get StatsGRPC all", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      map[int64]*relationgrpc.StatReply
			err      error
			mids     []int64
		)
		defer mockCtrl.Finish()
		mockArc := relationgrpc.NewMockRelationClient(mockCtrl)
		d.relGRPC = mockArc
		mockArc.EXPECT().Stats(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.StatsGRPC(ctx(), mids)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
