package native

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeRawNativePart(t *testing.T) {
	convey.Convey("RawNativePart", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativePart(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeRawNatPartIDs(t *testing.T) {
	convey.Convey("RawNatPartIDs", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NatPartIDsSearch(c, 1)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
