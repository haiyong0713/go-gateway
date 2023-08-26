package article

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-intl/interface/conf"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"
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

func TestArticles(t *testing.T) {
	Convey("Articles", t, func() {
		var (
			mockCtrl = gomock.NewController(t)
			res      map[int64]*article.Meta
			err      error
			aids     []int64
		)
		defer mockCtrl.Finish()
		mockArc := artclient.NewMockArticleGRPCClient(mockCtrl)
		d.artClient = mockArc
		mockArc.EXPECT().ArticleMetas(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.Articles(ctx(), aids)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
