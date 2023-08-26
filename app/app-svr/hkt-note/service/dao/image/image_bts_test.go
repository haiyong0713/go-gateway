package image

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestImage(t *testing.T) {
	c := context.Background()
	convey.Convey("AddImage", t, func(ctx convey.C) {
		res, err := d.Image(c, 1, 1)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestImages(t *testing.T) {
	c := context.Background()
	convey.Convey("Images", t, func(ctx convey.C) {
		res, err := d.Images(c, []int64{6, 7, 8, 9, 10}, 27515242)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
