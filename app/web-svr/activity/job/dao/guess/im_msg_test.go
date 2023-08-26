package guess

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/job/model/guess"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_SendImMsg(t *testing.T) {
	convey.Convey("SendImMsg", t, func(ctx convey.C) {
		var (
			c         = context.Background()
			RecverIDs = []uint64{27515232}
		)
		p := &guess.ImMsgParam{SenderUID: 88895135, MsgType: 1, Content: "私信内容", RecverIDs: RecverIDs}
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			code, err := d.SendImMsg(c, p)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(code, convey.ShouldEqual, 0)
			})
		})
	})
}
