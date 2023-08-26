package favorite

import (
	"context"
	"encoding/json"

	"fmt"
	"git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-gateway/app/app-svr/app-job/job/model/show"

	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFavoriteGenMedialist(t *testing.T) {
	var (
		ctx   = context.Background()
		serie = &show.Serie{}
		cover = ""
	)
	convey.Convey("GenMedialist", t, func(c convey.C) {
		fid, err := d.GenMedialist(ctx, serie, cover)
		c.Convey("Then err should be nil.fid should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(fid, convey.ShouldNotBeNil)
		})
	})
}

func TestFavoriteGenMedialist2(t *testing.T) {
	var (
		ctx = context.Background()
	)
	convey.Convey("GenMedialist", t, func(c convey.C) {
		fid, err := d.favClient.Favorites(ctx, &api.FavoritesReq{Fid: 352, Mid: 412466388, Pn: 1, Ps: 20})
		qq, _ := json.Marshal(fid)
		fmt.Println(string(qq))
		c.Convey("Then err should be nil.fid should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(fid, convey.ShouldNotBeNil)
		})
	})
}

func TestFavoriteAddMedias(t *testing.T) {
	var (
		ctx = context.Background()
		fid = int64(0)
		rsc = []*show.SerieRes{}
	)
	convey.Convey("AddMedias", t, func(c convey.C) {
		err := d.AddMedias(ctx, fid, rsc)
		c.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestFavoriteReplaceMedias(t *testing.T) {
	var (
		ctx = context.Background()
		fid = int64(0)
		rsc = []int64{1, 2, 3}
	)
	convey.Convey("ReplaceMedias", t, func(c convey.C) {
		err := d.ReplaceMedias(ctx, 0, fid, rsc)
		c.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
