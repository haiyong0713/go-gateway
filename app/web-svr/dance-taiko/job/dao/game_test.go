package dao

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/glycerine/goconvey/convey"

	"go-gateway/app/web-svr/dance-taiko/job/model"
)

func TestDao_AddCacheGame(t *testing.T) {
	var (
		c    = context.Background()
		game = &model.OttGame{
			GameId: 4,
			Aid:    10111848,
			Cid:    10152663,
			Status: model.GameJoining,
		}
	)
	convey.Convey("TestDao_AddCacheGame", t, func(ctx convey.C) {
		err := d.AddCacheGame(c, game)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_AddCachePlayers(t *testing.T) {
	var (
		c       = context.Background()
		gameId  = int64(4)
		players = make([]model.PlayerHonor, 0)
	)
	players = append(players, model.PlayerHonor{Mid: 123, Score: 2100})
	players = append(players, model.PlayerHonor{Mid: 321, Score: 2301})
	convey.Convey("TestDao_AddCachePlayers", t, func(ctx convey.C) {
		err := d.AddCachePlayers(c, gameId, players)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_RawPlayers(t *testing.T) {
	var (
		c      = context.Background()
		gameId = int64(5)
	)
	convey.Convey("TestDao_RawPlayers", t, func(ctx convey.C) {
		res, err := d.RawPlayers(c, gameId)
		str, _ := json.Marshal(res)
		convey.Println(string(str))
		ctx.So(res, convey.ShouldNotBeNil)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_UpdatePlayers(t *testing.T) {
	var (
		c       = context.Background()
		gameId  = int64(4)
		players = make([]*model.PlayerHonor, 0)
	)
	players = append(players, &model.PlayerHonor{Mid: 88895133, Score: 21001})
	players = append(players, &model.PlayerHonor{Mid: 1111120159, Score: 4444})
	convey.Convey("TestDao_UpdatePlayers", t, func(ctx convey.C) {
		err := d.UpdatePlayers(c, gameId, players)
		ctx.So(err, convey.ShouldBeNil)
	})
}
