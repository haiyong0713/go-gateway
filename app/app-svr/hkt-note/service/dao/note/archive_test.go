package note

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestViewPage(t *testing.T) {
	c := context.Background()
	convey.Convey("ViewPage", t, func(ctx convey.C) {
		res, err := d.ViewPage(c, 720019004)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArcs(t *testing.T) {
	c := context.Background()
	convey.Convey("ViewPage", t, func(ctx convey.C) {
		res, err := d.Arcs(c, []int64{640088483, 200063935, 320067800, 600089234, 600024226, 10113302})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSimpleArcs(t *testing.T) {
	c := context.Background()
	convey.Convey("SimpleArcs", t, func(ctx convey.C) {
		res, err := d.SimpleArcs(c, []int64{640088483, 200063935, 320067800, 600089234, 600024226, 10113302})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestCheeseSeasons(t *testing.T) {
	c := context.Background()
	convey.Convey("CheeseSeasons", t, func(ctx convey.C) {
		res, err := d.CheeseSeasons(c, []int32{81, 147})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSeasonEp(t *testing.T) {
	c := context.Background()
	convey.Convey("seasonEp", t, func(ctx convey.C) {
		res, total, err := d.seasonEp(c, 81, 1)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(total, convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestSeasonEps(t *testing.T) {
	c := context.Background()
	convey.Convey("seasonEps", t, func(ctx convey.C) {
		res, err := d.SeasonEps(c, 81)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestFeatureContList(t *testing.T) {
	c := context.Background()
	convey.Convey("FeatureContList", t, func(ctx convey.C) {
		res, err := d.FeatureContList(c)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
