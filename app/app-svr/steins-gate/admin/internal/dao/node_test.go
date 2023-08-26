package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoNode(t *testing.T) {
	convey.Convey("Node", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			key = int64(5621)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Node(c, key)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
