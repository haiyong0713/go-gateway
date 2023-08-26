package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoEpContests(t *testing.T) {
	var (
		c    = context.Background()
		keys = []int64{1, 2, 3}
	)
	convey.Convey("EpContests", t, func(ctx convey.C) {
		res, err := d.EpContests(c, keys)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoEpSeasons(t *testing.T) {
	var (
		c    = context.Background()
		keys = []int64{1, 2, 3}
	)
	convey.Convey("EpSeasons", t, func(ctx convey.C) {
		res, err := d.EpSeasons(c, keys)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoEpTeams(t *testing.T) {
	convey.Convey("EpTeams", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			keys = []int64{1, 2, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.EpTeams(c, keys)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchMainIDs(t *testing.T) {
	convey.Convey("SearchMainIDs", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.SearchMainIDs(c)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_SearchMD(t *testing.T) {
	convey.Convey("SearchMD", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			mainIDs = []int64{1, 2, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.SearchMD(c, mainIDs)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_EpGames(t *testing.T) {
	convey.Convey("EpGames", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			gids = []int64{1, 2, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.EpGames(c, gids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDao_EpGameMap(t *testing.T) {
	convey.Convey("EpGames", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			cids       = []int64{1, 5}
			tp   int64 = 3
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.EpGameMap(c, cids, tp)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}
