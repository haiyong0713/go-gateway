package lottery

import (
	"context"
	"testing"

	lott "go-gateway/app/web-svr/activity/interface/model/lottery"

	"github.com/smartystreets/goconvey/convey"
)

func TestLotterygetMsgKey(t *testing.T) {
	convey.Convey("getMsgKey", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			uids = []uint64{27515241}
			l    = &lott.LetterParam{RecverIDs: uids, SenderUID: uint64(37090048), MsgType: int32(1), Content: "测试"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, code, err := d.getMsgKey(c, l)
			convCtx.Convey("Then err should be nil.res,code should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(code, convey.ShouldNotBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotterySendLetter(t *testing.T) {
	convey.Convey("SendLetter", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			uids = []uint64{27515240}
			l    = &lott.LetterParam{RecverIDs: uids, SenderUID: uint64(37090048), MsgType: int32(1), Content: "恭喜您在抽奖活动中获得大会员优惠券，系统已自动发放到您的账户中，点击购买>> https://account.bilibili.com/account/big"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			code, err := d.SendLetter(c, l)
			convCtx.Convey("Then err should be nil.code should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(code, convey.ShouldNotBeNil)
			})
		})
	})
}
