package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoSendMessage(t *testing.T) {
	var (
		c     = context.Background()
		mids  = []int64{27515241}
		mc    = "1_24_1"
		title = "test"
		msg   = "testMsg"
	)
	convey.Convey("SendMessage", t, func(ctx convey.C) {
		err := d.SendMessage(c, mids, mc, title, msg)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
