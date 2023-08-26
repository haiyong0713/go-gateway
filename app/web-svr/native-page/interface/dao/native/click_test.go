package native

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeClicks(t *testing.T) {
	convey.Convey("Clicks", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.Clicks(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeRawNativeClicks(t *testing.T) {
	convey.Convey("RawNativeClicks", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativeClicks(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
