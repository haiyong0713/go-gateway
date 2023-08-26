package native

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawNatMixIDs(t *testing.T) {
	convey.Convey("RawNatMixIDs", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NatMixIDsSearch(c, 1, 1)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawNativeMixtures(t *testing.T) {
	convey.Convey("RawNativeMixtures", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativeMixtures(c, []int64{1})
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
