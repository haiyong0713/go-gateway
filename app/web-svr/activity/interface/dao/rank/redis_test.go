package rank

import (
	"context"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestGetRank(t *testing.T) {
	convey.Convey("GetRank", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetRank(c, "handWrite")
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
