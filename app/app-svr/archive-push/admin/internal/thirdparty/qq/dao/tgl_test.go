package dao

import (
	"flag"
	"github.com/glycerine/goconvey/convey"
	"go-common/library/conf/paladin"
	"go-common/library/log"
	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"testing"
)

func init() {
	flag.Set("conf", "/Users/zhouhaotian/Projects/go-gateway/app/app-svr/archive-push/admin/configs")
	flag.Set("deploy.env", "uat")
	log.Init(nil)
	paladin.Init()
	testD, _, _ = Init()
}

func Test_GetAccessToken(t *testing.T) {
	convey.Convey("Get access_token", t, func() {
		var (
			err   error
			token string
		)

		token, err = testD.GetAccessToken()
		convey.So(err, convey.ShouldBeNil)
		convey.So(token, convey.ShouldNotBeBlank)
	})
}

func Test_ContributeVideo(t *testing.T) {
	convey.Convey("contribute video", t, func() {
		var (
			err error
			res *qqModel.ContributeVideoReply
		)

		req := &qqModel.ContributeVideoReq{
			Title:     "bilibili测试稿件标题1",
			Summary:   "bilibili测试稿件简介",
			Cover:     "https://i0.hdslb.com/bfs/archive/67399a7f22e71849dd60b1b347e61890fb073585.jpg@320w_200h.jpg",
			Author:    "无敌皮皮鸟",
			Duration:  30,
			OuterVID:  "BV1Xv411x77y",
			OuterUser: "15484112",
			ExtTags:   "游戏,王者荣耀",
		}

		res, err = testD.ContributeVideo(req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldNotBeNil)
	})
}
