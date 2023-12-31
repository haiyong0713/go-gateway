package share

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestShares(t *testing.T) {
	convey.Convey("Shares", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.Shares(context.Background(), 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
