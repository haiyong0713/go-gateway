package steins

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoRawEdgeGroups(t *testing.T) {
	var (
		c   = context.Background()
		ids = []int64{}
	)
	convey.Convey("RawEdgeGroups", t, func(ctx convey.C) {
		res, err := d.RawEdgeGroups(c, ids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
