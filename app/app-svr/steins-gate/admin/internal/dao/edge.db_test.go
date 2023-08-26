package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoGraphEdgeList(t *testing.T) {
	convey.Convey("GraphEdgeList", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			graphid = int64(5621)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GraphEdgeList(c, graphid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoEdgeByNode(t *testing.T) {
	convey.Convey("EdgeByNode", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			fromNode = int64(5621)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.EdgeByNode(c, fromNode)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
