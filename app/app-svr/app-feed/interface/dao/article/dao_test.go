package article

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

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

func ctx() context.Context {
	return context.Background()
}

func TestArticles(t *testing.T) {
	Convey("Articles", t, func() {
		var (
			mockCtrl = gomock.NewController(t) // gomock来自github.com/otokaze/mock/gomock包
			res      map[int64]*article.Meta
			err      error
		)
		defer mockCtrl.Finish()
		mockArc := artclient.NewMockArticleGRPCClient(mockCtrl) // mock来自由creater生成的go-common/app/service/main/archive/api/grpc/v1/mock包
		d.artClient = mockArc
		mockArc.EXPECT().ArticleMetas(ctx(), gomock.Any()).Return(res, nil)
		res, err = d.Articles(ctx(), []int64{111005921})
		So(res, ShouldNotBeNil)
		So(err, ShouldNotBeNil)
	})
}
