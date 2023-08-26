package bnj

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestComicCoupon(t *testing.T) {
	convey.Convey("ComicCoupon", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(2089809)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.ComicCoupon(c, mid, 1)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
