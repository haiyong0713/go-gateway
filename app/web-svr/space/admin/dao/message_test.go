package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoSendSystemMessage(t *testing.T) {
	convey.Convey("SendSystemMessage", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mids  = []int64{27515256}
			mc    = "1_26_1"
			title = "测试系统消息"
			msg   = "测试系统消息内容"
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SendSystemMessage(c, mids, mc, title, msg)
			ctx.Convey("Then err should be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
