package dao

import (
	"fmt"
	"github.com/glycerine/goconvey/convey"
	bzModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
	"testing"
)

func Test_VodAddReq_PushUp(t *testing.T) {
	convey.Convey("VodAdd PushUp", t, func() {
		var (
			err error
			res *bzModel.VodAddReply
		)

		testReq := bzModel.VodAddReq{
			BVID:        "BV1nr4y1K7gS",
			Page:        1,
			Category:    bzModel.DefaultVodAddCategory,
			Title:       "【风暴英雄】滑稽时刻第77期：伤害不高，侮辱极强~",
			Description: "相关游戏：风暴英雄\n投稿邮箱：794917828@qq.com （注明剧情和时间点） \n爱发电主页：https://afdian.net/@Army95\npatreon主页：https://www.patreon.com/Army95\n油管地址：Youtube/Army军.com\n网易云音乐：Army菌 \n微博：Army菌 \nB 站：Army菌\nQQ群：390724504    \n游戏频道：/join滑稽时刻\n战网社交群组：747ejhav5",
			Stage:       "风暴英雄",
			Duration:    564,
			Thumbnail:   "https://i2.hdslb.com/bfs/archive/6411810126ff3b0575f87e31ccde915c7dd20ef9.jpg@320w_200h.jpg",
			Status:      bzModel.VodAddStatusPushUp,
		}

		res, err = testD.VodAdd(testReq)
		if err != nil {
			fmt.Printf("error: %v", err)
		} else {
			fmt.Printf("response: %v", res)
		}
	})

	convey.Convey("VodAdd PushUp", t, func() {
		var (
			err error
			res *bzModel.VodAddReply
		)

		testReq := bzModel.VodAddReq{
			BVID:        "BV12E411t7w3",
			Page:        1,
			Category:    bzModel.DefaultVodAddCategory,
			Title:       "【星际老男孩】11月2号世界锦标赛全球总决赛决赛日",
			Description: "(=・ω・=)视频我检查过大致没问题，如果哪里有什么问题记得艾特我。另外说一下，视频文件足有28G大小，所以没办法按照某些人的要求来剪切。如果要花时间去合并并且剪切的话，之后还要调整音轨，那你们估计要在两天以后才能看到这个视频。",
			Stage:       "星际2",
			Duration:    6501,
			Thumbnail:   "https://i1.hdslb.com/bfs/archive/91392b367aea1d7b0a0b07da2d28c2d7aa08329b.png@320w_200h.png",
			Status:      bzModel.VodAddStatusPushUp,
		}

		res, err = testD.VodAdd(testReq)
		if err != nil {
			fmt.Printf("error: %v", err)
		} else {
			fmt.Printf("response: %v", res)
		}
	})
}

func Test_VodAddReq_Withdraw(t *testing.T) {

	convey.Convey("VodAdd Withdraw", t, func() {
		var (
			err error
			res *bzModel.VodAddReply
		)

		testReq := bzModel.VodAddReq{
			BVID:        "BV1st41127iU",
			Page:        1,
			Category:    bzModel.DefaultVodAddCategory,
			Title:       "稳了！这才暴雪有史以来最明智的决策！给世界粉丝们的最大惊喜！必将走向癫疯",
			Description: "暴雪的游戏重制了，我的视频也重制了！\n由于上次个人原因，视频内容出现漏洞！给很多玩家造成了不良影响！\n对此，我表示很抱歉！\n请你们务必看看这次我重制过的视频！\n谢谢了，感恩！",
			Stage:       "风暴英雄",
			Duration:    358,
			Thumbnail:   "https://i1.hdslb.com/bfs/archive/3d18a427ab37a145ef188b6aa5141eda64e5c9c3.jpg@320w_200h.jpg",
			Status:      bzModel.VodAddStatusWithdraw,
		}

		res, err = testD.VodAdd(testReq)
		if err != nil {
			fmt.Printf("error: %v", err)
		} else {
			fmt.Printf("response: %v", res)
		}
	})
}
