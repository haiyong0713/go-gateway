package location

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
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
	if os.Getenv("UT_LOCAL_TEST") != "" {
		dir, _ := filepath.Abs("../../cmd/app-feed-test.toml")
		flag.Set("conf", dir)
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestInfo(t *testing.T) {
	Convey("get Info", t, func() {
		var (
			ipaddr   string
			mockCtrl = gomock.NewController(t)
			err      error
			res      *locgrpc.InfoReply
		)
		defer mockCtrl.Finish()
		mockArc := locgrpc.NewMockLocationClient(mockCtrl)
		d.locGRPC = mockArc
		mockArc.EXPECT().Info(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.InfoGRPC(ctx(), ipaddr)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
