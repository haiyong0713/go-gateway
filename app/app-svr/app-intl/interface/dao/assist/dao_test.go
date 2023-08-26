package assist

import (
	"context"
	"flag"
	"os"
	"testing"

	assistApi "git.bilibili.co/bapis/bapis-go/assist/service"
	"go-gateway/app/app-svr/app-intl/interface/conf"

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

func TestAssist(t *testing.T) {
	Convey("Articles", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			err      error
			res      []int64
			upMid    int64
		)
		defer mockCtrl.Finish()
		mockArc := assistApi.NewMockAssistClient(mockCtrl)
		d.assistGRPC = mockArc
		mockArc.EXPECT().AssistIDs(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.Assist(ctx(), upMid)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
