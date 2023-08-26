package videoup

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	vuapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-gateway/app/app-svr/app-intl/interface/conf"

	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func init() {
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
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestArcViewAddit(t *testing.T) {
	Convey("TestArcViewAddit", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      *vuapi.ArcViewAdditReply
			err      error
			aids     int64
		)
		defer mockCtrl.Finish()
		mockArc := vuapi.NewMockVideoUpOpenClient(mockCtrl)
		d.videoupGRPC = mockArc
		mockArc.EXPECT().ArcViewAddit(context.Background(), gomock.Any()).Return(res, nil)
		res, err = d.ArcViewAddit(context.Background(), aids)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
