package result

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestResultSnAids(t *testing.T) {
	convey.Convey("SnAids", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			aidsToSns, _, err := d.SnAids(c)
			convCtx.Convey("Then err should be nil.aidsToSns should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(aidsToSns, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestResultSnArcs(t *testing.T) {
	convey.Convey("SnArcs", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, aids, err := d.SnArcs(c, sid)
			convCtx.Convey("Then err should be nil.res,aids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(aids, convey.ShouldNotBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
