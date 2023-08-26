package location

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-intl/interface/conf"

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
		flag.Set("app_id", "main.app-svr.app-intl")
		flag.Set("conf_token", "02007e8d0f77d31baee89acb5ce6d3ac")
		flag.Set("tree_id", "64518")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-intl-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
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
		res, err = d.Info(ctx(), ipaddr)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestAuthPIDs(t *testing.T) {
	Convey("auth", t, func() {
		var (
			pids, ipaddr string
			mockCtrl     = gomock.NewController(t)
			err          error
			res          map[string]*locgrpc.Auth
		)
		defer mockCtrl.Finish()
		mockArc := locgrpc.NewMockLocationClient(mockCtrl)
		d.locGRPC = mockArc
		mockArc.EXPECT().AuthPIDs(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.AuthPIDs(ctx(), pids, ipaddr)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestArchive(t *testing.T) {
	Convey("Archive", t, func() {
		var (
			aid, mid      int64
			ipaddr, cndip string
			mockCtrl      = gomock.NewController(t)
			err           error
			res           *locgrpc.Auth
		)
		defer mockCtrl.Finish()
		mockArc := locgrpc.NewMockLocationClient(mockCtrl)
		d.locGRPC = mockArc
		mockArc.EXPECT().Archive(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.Archive(ctx(), aid, mid, ipaddr, cndip)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
