package ott

import (
	"context"
	"fmt"
	"github.com/glycerine/goconvey/convey"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
	"testing"
)

func TestDao_SelectGamesByCid(t *testing.T) {
	var (
		c   = context.Background()
		cid = int64(10099306)
	)
	convey.Convey("TestDao_SelectGamesByCid", t, func(ctx convey.C) {
		res, err := d.SelectGamesByCid(c, cid)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_CreateGame(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10099306)
		cid = int64(10109083)
	)
	convey.Convey("TestDao_CreateGame", t, func(ctx convey.C) {
		res, err := d.CreateGame(c, aid, cid)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
		fmt.Println(res)
	})
}

func TestDao_StartGame(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(1)
	)
	convey.Convey("TestDao_StartGame", t, func(ctx convey.C) {
		err := d.StartGame(c, gameId)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_RawGame(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(1)
	)
	convey.Convey("TestDao_RawGame", t, func(ctx convey.C) {
		res, err := d.rawGame(c, gameId)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_LoadGame(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(5)
	)
	convey.Convey("TestDao_LoadGame", t, func(ctx convey.C) {
		res, err := d.LoadGame(c, gameId)
		fmt.Printf("%+v", res)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_AddCacheGamePkg(t *testing.T) {
	var (
		c   = context.Background()
		url = "testtest"
	)
	convey.Convey("TestDao_AddCacheGamePkg", t, func(ctx convey.C) {
		err := d.AddCacheGamePkg(c, url)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_CacheGamePkg(t *testing.T) {
	var c = context.Background()
	convey.Convey("TestDao_CacheGamePkg", t, func(ctx convey.C) {
		res, err := d.CacheGamePkg(c)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_addCacheGame(t *testing.T) {
	var (
		c    = context.Background()
		game = &model.OttGame{Aid: 123, Cid: 123, GameId: 4}
	)
	convey.Convey("TestDao_addCacheGame", t, func(ctx convey.C) {
		err := d.addCacheGame(c, game.GameId, game)
		ctx.So(err, convey.ShouldBeNil)
	})

}

func TestDao_DelCaches(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(150)
		mids   = []int64{123}
	)
	convey.Convey("TestDao_DelCaches", t, func(ctx convey.C) {
		err := d.DelCaches(c, gameId, mids)
		ctx.So(err, convey.ShouldBeNil)
	})
}
