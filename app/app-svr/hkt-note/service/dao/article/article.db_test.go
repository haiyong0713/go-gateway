package article

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/interface/model/article"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawArtDetails(t *testing.T) {
	c := context.Background()
	convey.Convey("rawArtDetails", t, func(ctx convey.C) {
		res, err := d.rawArtDetails(c, []int64{4543}, article.TpArtDetailCvid, 2, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestRawArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("RawArtDetail", t, func(ctx convey.C) {
		res, err := d.RawArtDetail(c, 4524, article.TpArtDetailCvid, 2, 4)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestRawArtListInUser(t *testing.T) {
	c := context.Background()
	convey.Convey("rawArtListInUser", t, func(ctx convey.C) {
		res, _, err := d.rawArtListInUser(c, 0, -1, 27515251)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestRawArtListInArc(t *testing.T) {
	c := context.Background()
	convey.Convey("rawArtListInUser", t, func(ctx convey.C) {
		res, cvids, err := d.rawArtListInArc(c, 0, -1, 10113209, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
			ctx.So(len(cvids), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestRawArtContent(t *testing.T) {
	c := context.Background()
	convey.Convey("rawArtContent", t, func(ctx convey.C) {
		res, err := d.rawArtContent(c, 1, 1)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestRawArtCountInArc(t *testing.T) {
	c := context.Background()
	convey.Convey("artCountInArc", t, func(ctx convey.C) {
		res, err := d.rawArtCountInArc(c, 320009175, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestFilterLockArts(t *testing.T) {
	c := context.Background()
	convey.Convey("filterLockArts", t, func(ctx convey.C) {
		res, err := d.filterLockArts(c, []int64{4543, 4542})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
