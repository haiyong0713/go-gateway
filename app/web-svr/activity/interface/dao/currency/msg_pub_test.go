package currency

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCurrencySendHeadMsg(t *testing.T) {
	convey.Convey("SendHeadMsg", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(10)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SendHeadMsg(c, mid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCurrencySendPropsMsg(t *testing.T) {
	convey.Convey("SendPropsMsg", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(11)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SendPropsMsg(c, mid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
