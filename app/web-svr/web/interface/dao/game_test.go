package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestDaoGameInfo(t *testing.T) {
	convey.Convey("GameInfo", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			id int64 = 28
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.gameInfoURL).Reply(200).JSON(`{"code":0,"data":{"game_name":"1111","game_icon":"1111"}}`)
			data, err := d.GameInfo(c, id)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestDaoGameSearchInfo(t *testing.T) {
	convey.Convey("GameSearchInfo", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			id int64 = 85
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.searchGameInfoURL).Reply(200).JSON(`{"code":0,"data":{"game_name":"1111","game_icon":"1111"}}`)
			data, err := d.SearchGameInfo(c, id)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}
