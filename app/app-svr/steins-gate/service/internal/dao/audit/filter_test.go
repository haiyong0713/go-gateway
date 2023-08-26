package audit

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoMFilterMsg(t *testing.T) {
	convey.Convey("MFilterMsg", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			area = "steins_gate"
			msgs = map[string]string{"test": "test"}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			data, err := d.MFilterMsg(c, area, msgs)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoFilterMsg(t *testing.T) {
	convey.Convey("FilterMsg", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			area = "steins_gate"
			msg  = "测试"
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			data, err := d.FilterMsg(c, area, msg)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}
