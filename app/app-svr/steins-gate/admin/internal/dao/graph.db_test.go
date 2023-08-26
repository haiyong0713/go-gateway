package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoGraphAuditByID(t *testing.T) {
	convey.Convey("GraphAuditByID", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			graphid = int64(5621)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GraphAuditByID(c, graphid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
