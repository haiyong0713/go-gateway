package dao

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/glycerine/goconvey/convey"
)

func TestDao_BwsCreateRoom(t *testing.T) {
	var c = context.Background()
	convey.Convey("TestDao_BwsCreateRoom", t, func(ctx convey.C) {
		res, err := d.BwsCreateRoom(c)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}

func TestDao_BwsMidInfo(t *testing.T) {
	var (
		c      = context.Background()
		gameId = 102
		mid    = 27515256
	)
	convey.Convey("TestDao_BwsMidInfo", t, func(c2 convey.C) {
		res, err := d.BwsMidInfo(c, int64(mid), gameId)
		c2.So(err, convey.ShouldBeNil)
		c2.So(res.Valid, convey.ShouldBeTrue)
		fmt.Println(res)
	})
}

func TestDao_BwsJoinRoom(t *testing.T) {
	var (
		c      = context.Background()
		gameId = 102
		mid    = 27515256
	)
	convey.Convey("TestDao_BwsJoinRoom", t, func(c2 convey.C) {
		err := d.BwsJoinRoom(c, gameId, int64(mid))
		c2.So(err, convey.ShouldBeNil)
	})
}

func TestDao_BwsStartGame(t *testing.T) {
	var (
		c      = context.Background()
		gameId = 102
	)
	convey.Convey("TestDao_BwsStartGame", t, func(c2 convey.C) {
		err := d.BwsStartGame(c, gameId)
		c2.So(err, convey.ShouldBeNil)
	})
}

func TestDao_BwsEndGame(t *testing.T) {
	var (
		c      = context.Background()
		gameId = 102
		player = &model.BwsPlayResult{
			Mid:   27515256,
			Score: 1234,
			Star:  3,
		}
	)
	convey.Convey("TestDao_BwsEndGame", t, func(c2 convey.C) {
		err := d.BwsEndGame(c, gameId, []*model.BwsPlayResult{player})
		c2.So(err, convey.ShouldBeNil)
	})
}
