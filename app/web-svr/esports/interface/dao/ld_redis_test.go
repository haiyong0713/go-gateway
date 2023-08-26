package dao

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/esports/interface/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_CacheLolGames(t *testing.T) {
	convey.Convey("CacheLolGames", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			matchID = int64(524698)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheLolGames(c, matchID)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDao_AddCacheLolGames(t *testing.T) {
	convey.Convey("AddCacheLolGames", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			data    []*model.LolGame
			matchID = int64(524698)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.AddCacheLolGames(c, matchID, data)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheDotaGames(t *testing.T) {
	convey.Convey("CacheDotaGames", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			matchID = int64(502945)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheDotaGames(c, matchID)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDao_AddCacheDotaGames(t *testing.T) {
	convey.Convey("AddCacheDotaGames", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			data    []*model.LolGame
			matchID = int64(502945)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.AddCacheDotaGames(c, matchID, data)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheOwGames(t *testing.T) {
	convey.Convey("CacheOwGames", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			matchID = int64(542674)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheOwGames(c, matchID)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDao_AddCacheOwGames(t *testing.T) {
	convey.Convey("AddCacheOwGames", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			data    []*model.OwGame
			matchID = int64(542674)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.AddCacheOwGames(c, matchID, data)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
