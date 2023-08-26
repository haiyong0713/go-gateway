package guess

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_UserStat(t *testing.T) {
	var (
		c         = context.Background()
		mid       = int64(27515412)
		stakeType = int64(1)
		business  = int64(1)
	)
	convey.Convey("UserStat", t, func(ctx convey.C) {
		res, err := d.UserStat(c, mid, stakeType, business)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_GuessMain(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27515412)
	)
	convey.Convey("GuessMain", t, func(ctx convey.C) {
		res, err := d.GuessMain(c, mid)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_UserGuessList(t *testing.T) {
	var (
		c        = context.Background()
		mid      = int64(27515412)
		business = int64(1)
	)
	convey.Convey("UserGuessList", t, func(ctx convey.C) {
		res, err := d.UserGuessList(c, mid, business)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestDao_MDResult(t *testing.T) {
	var (
		c        = context.Background()
		mid      = int64(27515412)
		business = int64(1)
	)
	convey.Convey("MDResult", t, func(ctx convey.C) {
		res, err := d.MDResult(c, mid, business)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_MDsResult(t *testing.T) {
	var (
		c        = context.Background()
		mainIDs  = []int64{15555180}
		business = int64(1)
	)
	convey.Convey("MDsResult", t, func(ctx convey.C) {
		res, err := d.MDsResult(c, mainIDs, business)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.Print(res)
		})
	})
}

func TestDao_OidsMIDs(t *testing.T) {
	var (
		c        = context.Background()
		oids     = []int64{83}
		business = int64(1)
	)
	convey.Convey("OidsMIDs", t, func(ctx convey.C) {
		res, err := d.OidsMIDs(c, oids, business)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestDao_OidMIDs(t *testing.T) {
	var (
		c        = context.Background()
		oid      = int64(83)
		business = int64(1)
	)
	convey.Convey("OidMIDs", t, func(ctx convey.C) {
		res, err := d.OidMIDs(c, oid, business)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestDao_UserGuess(t *testing.T) {
	convey.Convey("UserGuess", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			mid     = int64(1)
			mainIDs = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UserGuess(c, mainIDs, mid)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
