package article

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestArtDetails(t *testing.T) {
	c := context.Background()
	convey.Convey("ArtDetails", t, func(ctx convey.C) {
		res, err := d.ArtDetails(c, []int64{9970684826484742, 10066914834907146, 9970684826484742}, "note_id")
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("ArtDetail", t, func(ctx convey.C) {
		res, err := d.ArtDetail(c, 4524, "cvid")
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArtCountInUser(t *testing.T) {
	c := context.Background()
	convey.Convey("ArtCountInUser", t, func(ctx convey.C) {
		res, err := d.ArtCountInUser(c, 27515242)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArtListInArc(t *testing.T) {
	c := context.Background()
	convey.Convey("ArtListInArc", t, func(ctx convey.C) {
		res, err := d.ArtListInArc(c, 0, -1, 680122230, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
