package newstar

import (
	"context"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestDao_JoinNewstar(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("JoinNewstar", t, func(ctx convey.C) {
		lastID, err := d.JoinNewstar(c, "newstar_first", 0, 88895033, 0)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(lastID, convey.ShouldBeGreaterThanOrEqualTo, 0)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_RawCreation(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RawCreation", t, func(ctx convey.C) {
		_, err := d.RawCreation(c, "newstar_first", 88895033)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_InviteCount(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("InviteCount", t, func(ctx convey.C) {
		count, err := d.InviteCount(c, "newstar_first", 88895033)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(count, convey.ShouldBeGreaterThanOrEqualTo, 0)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_RawInvites(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RawInvites", t, func(ctx convey.C) {
		res, err := d.RawInvites(c, "newstar_first", 88895033)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_RawAwards(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RawAwards", t, func(ctx convey.C) {
		res, err := d.RawAwards(c)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
