package favorite

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFavoritefavStype(t *testing.T) {
	var (
		sType = "weekly_selected"
	)
	convey.Convey("favStype", t, func(ctx convey.C) {
		oid := favStype(sType)
		ctx.Convey("Then oid should not be nil.", func(ctx convey.C) {
			ctx.So(oid, convey.ShouldNotBeNil)
		})
	})
}

func TestFavoriteFavAdd(t *testing.T) {
	var (
		ctx   = context.Background()
		mid   = int64(1)
		sType = "weekly_selected"
	)
	convey.Convey("FavAdd", t, func(c convey.C) {
		err := d.FavAdd(ctx, mid, sType)
		c.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestFavoriteFavDel(t *testing.T) {
	var (
		ctx   = context.Background()
		mid   = int64(1)
		sType = "weekly_selected"
	)
	convey.Convey("FavDel", t, func(c convey.C) {
		err := d.FavDel(ctx, mid, sType)
		c.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestFavoriteFavCheck(t *testing.T) {
	var (
		ctx   = context.Background()
		mid   = int64(1)
		sType = "weekly_selected"
	)
	convey.Convey("FavCheck", t, func(c convey.C) {
		fav, err := d.FavCheck(ctx, mid, sType)
		c.Convey("Then err should be nil.fav should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(fav, convey.ShouldNotBeNil)
		})
	})
}
