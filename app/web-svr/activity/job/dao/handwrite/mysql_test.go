package handwrite

import (
	"context"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestMidListDistinct(t *testing.T) {
	convey.Convey("MidListDistinct", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			mids = []int64{}
		)
		mids = append(mids, 2, 1, 3)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.MidListDistinct(c, mids)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
