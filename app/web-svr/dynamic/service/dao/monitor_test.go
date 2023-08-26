package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoSend(t *testing.T) {
	convey.Convey("Send", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			msg  = "test robot"
			urls = "http://bap.bilibili.co/api/v1/message/add"
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			httpMock("POST", urls).Reply(200).JSON(`{"status":400012}`)
			err := d.Send(c, msg)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
