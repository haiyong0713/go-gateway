package ott

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/glycerine/goconvey/convey"
)

func TestDao_AddCacheRanks(t *testing.T) {
	var (
		c       = context.Background()
		cid     = int64(10099306)
		players = make([]*model.PlayerHonor, 2)
	)
	players[0] = &model.PlayerHonor{Mid: 123, Score: 123}
	players[1] = &model.PlayerHonor{Mid: 213, Score: 321}
	convey.Convey("TestDao_AddCacheRanks", t, func(ctx convey.C) {
		err := d.AddCacheRanks(c, cid, players)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_CacheRank(t *testing.T) {
	var (
		c   = context.Background()
		cid = int64(10099306)
	)
	convey.Convey("TestDao_AddCacheRanks", t, func(ctx convey.C) {
		res, err := d.CacheRank(c, cid, 0, 20)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_AddCacheRank(t *testing.T) {
	var (
		c       = context.Background()
		cid     = int64(10099306)
		players = make([]*model.PlayerHonor, 3)
	)
	players[0] = &model.PlayerHonor{Mid: 123456, Score: 2000}
	players[1] = &model.PlayerHonor{Mid: 321456, Score: 2100}
	players[2] = &model.PlayerHonor{Mid: 654321, Score: 2500}
	convey.Convey("TestDao_AddCacheRank", t, func(ctx convey.C) {
		err := d.AddCacheRank(c, cid, players)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_CachePlayerRank(t *testing.T) {
	var (
		c   = context.Background()
		cid = int64(10152663)
		mid = []int64{27515401, 110000290, 2061885852}
	)
	convey.Convey("TestDao_CachePlayerRank", t, func(ctx convey.C) {
		res, err := d.CachePlayersRank(c, cid, mid)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_CachePlayerScore(t *testing.T) {
	var (
		c   = context.Background()
		cid = int64(10099306)
		mid = int64(123)
	)
	convey.Convey("TestDao_CachePlayerScore", t, func(ctx convey.C) {
		res, err := d.CachePlayerScore(c, cid, mid)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_LoadRanks(t *testing.T) {
	var (
		c   = context.Background()
		cid = int64(10152663)
	)
	convey.Convey("TestDao_LoadRanks", t, func(ctx convey.C) {
		res, err := d.LoadRanks(c, cid, 0, 10)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}
