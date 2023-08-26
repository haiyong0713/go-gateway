package image

import (
	"context"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAddImage(t *testing.T) {
	c := context.Background()
	convey.Convey("AddImage", t, func(ctx convey.C) {
		mid := int64(1)
		location := "/bfs/note/508322209e8dacc6830739a3447c41f1a110c651.jpg"
		id, err := d.AddImage(c, mid, location)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(id, convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestRawImage(t *testing.T) {
	c := context.Background()
	convey.Convey("Image", t, func(ctx convey.C) {
		res, err := d.rawImage(c, 27515242, 1)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res.Location), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestRawImages(t *testing.T) {
	c := context.Background()
	convey.Convey("Image", t, func(ctx convey.C) {
		res, err := d.rawImages(c, []int64{1, 2, 3}, 27515242)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}
