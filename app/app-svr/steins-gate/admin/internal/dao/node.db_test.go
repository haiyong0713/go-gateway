package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoGraphNodeList(t *testing.T) {
	convey.Convey("GraphNodeList", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			graphid = int64(5621)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GraphNodeList(c, graphid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoRawNode(t *testing.T) {
	convey.Convey("RawNode", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(5621)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawNode(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
