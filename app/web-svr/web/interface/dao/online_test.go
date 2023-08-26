package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoOnlineList(t *testing.T) {
	convey.Convey("OnlineList", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			num = int64(10)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.OnlineList(c, num)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoOnlineTotal(t *testing.T) {
	convey.Convey("OnlineTotal", t, func(ctx convey.C) {
		c := context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.OnlineTotal(c)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}
