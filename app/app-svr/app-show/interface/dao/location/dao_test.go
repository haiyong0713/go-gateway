package location

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

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
		flag.Set("app_id", "main.app-svr.app-show")
		flag.Set("conf_token", "Pae4IDOeht4cHXCdOkay7sKeQwHxKOLA")
		flag.Set("tree_id", "2687")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-show-test.toml")

	}
	flag.Parse()
	cfg, err := confInit()
	if err != nil {
		panic(err)
	}
	d = New(cfg)
	os.Exit(m.Run())
}

func confInit() (*conf.Config, error) {
	err := paladin.Init()
	if err != nil {
		return nil, err
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("app-show.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func TestInfo(t *testing.T) {
	Convey("get Info", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      *locgrpc.InfoReply
			err      error
		)
		defer mockCtrl.Finish()
		mockArc := locgrpc.NewMockLocationClient(mockCtrl)
		d.locGRPC = mockArc
		mockArc.EXPECT().Info(context.TODO(), gomock.Any()).Return(res, nil)
		res, err = d.Info(ctx(), "127.0.0.1")
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestAuthPIDs(t *testing.T) {
	Convey("get AuthPIDs", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      map[string]*locgrpc.Auth
			err      error
		)
		defer mockCtrl.Finish()
		mockArc := locgrpc.NewMockLocationClient(mockCtrl)
		d.locGRPC = mockArc
		mockArc.EXPECT().AuthPIDs(context.TODO(), gomock.Any()).Return(res, nil)
		res, err = d.AuthPIDs(ctx(), "417,1521", "127.0.0.0")
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
