package resource

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestResourceDySear(t *testing.T) {
	convey.Convey("DySearch", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.DySearch(c)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
