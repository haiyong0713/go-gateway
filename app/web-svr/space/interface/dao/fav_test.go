package dao

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/space/interface/model"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestDao_AlbumFavCount(t *testing.T) {
	convey.Convey("test album fav count", t, func(ctx convey.C) {
		defer gock.OffAll()
		httpMock("GET", d.favAlbumURL).Reply(200).JSON(`{"code": 0, "data":{"pageinfo":{"count":1}}}`)
		mid := int64(88895029)
		favType := 2
		data, err := d.LiveFavCount(context.Background(), mid, favType)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%v", data)
	})
}

func TestDao_FavArchive(t *testing.T) {
	convey.Convey("test fav archive", t, func(ctx convey.C) {
		defer gock.OffAll()
		httpMock("GET", d.favArcURL).Reply(200).JSON(`{"code": 0}`)
		mid := int64(88895029)
		arg := &model.FavArcArg{}
		data, err := d.FavArchive(context.Background(), mid, arg)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%v", data)
	})
}
