package steins

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoEdgeGroups(t *testing.T) {
	convey.Convey("GraphShow", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			IDs = []int64{77, 88}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.EdgeGroups(c, IDs)
			convCtx.Convey("Then err should be nil.data should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}
