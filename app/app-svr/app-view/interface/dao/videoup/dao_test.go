package videoup

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model/view"

	vuapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
)

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-view")
		flag.Set("conf_token", "3a4CNLBhdFbRQPs7B4QftGvXHtJo92xw")
		flag.Set("tree_id", "4575")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-view-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	dao = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestMaterialView(t *testing.T) {
	var (
		c      = context.Background()
		params = &view.MaterialParam{
			AID:      1,
			CID:      1,
			Build:    8470,
			Platform: "ios",
			Device:   "phone",
			MobiApp:  "iphone",
		}
	)
	convey.Convey("MaterialView", t, func(ctx convey.C) {
		material, err := dao.MaterialView(c, params)
		fmt.Printf("-----%+v-------", material)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArcCommercial(t *testing.T) {
	convey.Convey("ArcCommercial", t, func() {
		var (
			c         = context.Background()
			aid int64 = 10113235
			err error
			res = &vuapi.ArcCommercialReply{
				GameID: 1,
			}
		)
		convey.Convey("Then res should not be nil", func() {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()
			mockClient := vuapi.NewMockVideoUpOpenClient(mockCtl)
			dao.videoupGRPC = mockClient
			mockClient.EXPECT().ArcCommercial(c, gomock.Any()).Return(res, nil)
			res.GameID, err = dao.ArcCommercial(c, aid)
			fmt.Println("res.GameId:", res.GameID)
			convey.So(res.GameID, convey.ShouldNotBeZeroValue)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArcBgmList(t *testing.T) {
	convey.Convey("ArcBgmList", t, func() {
		var (
			c         = context.Background()
			aid int64 = 10200334
			cid int64 = 10167155
		)
		bgms := make([]*vuapi.ViewBGM, 1)
		bgms[0] = &vuapi.ViewBGM{Sid: 1, Mid: 1, Title: "Mock", Author: "Mock", JumpUrl: "www.bilibili.com", Cover: "mock"}
		res := &vuapi.BgmListReply{
			Bgms: bgms,
		}
		convey.Convey("Then res should not be empty", func() {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()
			mockClient := vuapi.NewMockVideoUpOpenClient(mockCtl)
			dao.videoupGRPC = mockClient
			mockClient.EXPECT().ArcBgmList(c, gomock.Any()).Return(res, nil)
			tmp, err := dao.ArcBgmList(c, aid, cid)
			fmt.Println(tmp)
			convey.So(res.Bgms, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArcViewAddit(t *testing.T) {
	var (
		c         = context.Background()
		aid int64 = 880035114
	)
	convey.Convey("TestArcViewAddit", t, func(ctx convey.C) {
		addit, err := dao.ArcViewAddit(c, aid)
		fmt.Printf("-----%+v-------", addit)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetMaterialList(t *testing.T) {
	var (
		c         = context.Background()
		aid int64 = 560008223
		cid int64 = 10300412
	)
	convey.Convey("TestGetMaterialList", t, func(ctx convey.C) {
		res, _, _, err := dao.GetMaterialList(c, aid, cid)
		fmt.Printf("-----%+v-------", res)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
