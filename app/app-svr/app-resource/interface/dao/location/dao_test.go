package location

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/location"

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

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-resource")
		flag.Set("conf_token", "z8JNX5MFIyDxyBsqwQyF6pnjWQ5YOA14")
		flag.Set("tree_id", "2722")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-resource-test.toml")
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
			mockCtrl = gomock.NewController(t)
			res      *location.Info
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
			res      map[int64]*locgrpc.Auth
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

func TestInfoComplete(t *testing.T) {
	Convey("InfoComplete", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      *locgrpc.InfoComplete
			err      error
		)
		defer mockCtrl.Finish()
		mockArc := locgrpc.NewMockLocationClient(mockCtrl)
		d.locGRPC = mockArc
		mockArc.EXPECT().InfoComplete(context.TODO(), gomock.Any()).Return(res, nil)
		res, err = d.InfoComplete(ctx(), "127.0.0.1")
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestZlimitInfo(t *testing.T) {
	Convey("ZlimitInfo", t, func() {
		res, err := d.ZlimitInfo(ctx(), []string{"24.51.1.227"}, []int64{2140})
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
		fmt.Printf("%v", res)
	})
}
