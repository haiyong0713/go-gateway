package ott

import (
	"context"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestDao_AddPlayer(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(1)
		mid    = int64(20589727)
	)
	convey.Convey("TestDao_AddPlayer", t, func(ctx convey.C) {
		err := d.AddPlayer(c, gameId, mid)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_rawPlayers(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(1)
	)
	convey.Convey("TestDao_rawPlayers", t, func(ctx convey.C) {
		res, err := d.RawPlayers(c, gameId)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_SelectPlayersByGames(t *testing.T) {
	var (
		c   = context.Background()
		ids = []int64{1}
	)
	convey.Convey("TestDao_SelectPlayersByGames", t, func(ctx convey.C) {
		res, err := d.SelectPlayersByGames(c, ids)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_CachePlayersCombo(t *testing.T) {
	var (
		c      = context.Background()
		mids   = []int64{1}
		gameId = int64(1)
	)
	convey.Convey("TestDao_CachePlayersCombo", t, func(ctx convey.C) {
		res, err := d.CachePlayersCombo(c, gameId, mids)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_LoadPlayers(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(3)
	)
	convey.Convey("TestDao_LoadPlayers", t, func(ctx convey.C) {
		res, err := d.LoadPlayers(c, gameId)
		ctx.So(res, convey.ShouldNotBeNil)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_CachePlayer(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(3)
	)
	convey.Convey("TestDao_CachePlayer", t, func(ctx convey.C) {
		res, err := d.CachePlayer(c, gameId)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}
