package native

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawNativeTab(t *testing.T) {
	convey.Convey("RawNativeTab", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativeTabs(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
