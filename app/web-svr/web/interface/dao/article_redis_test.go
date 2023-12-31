package dao

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/web/interface/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaokeyArtList(t *testing.T) {
	convey.Convey("keyArtList", t, func(ctx convey.C) {
		var (
			rid  = int64(0)
			sort = int32(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyArtList(rid, sort)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoArticleUpListCache(t *testing.T) {
	convey.Convey("ArticleUpListCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ArticleUpListCache(c)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", res)
			})
		})
	})
}

func TestDaoSetArticleUpListCache(t *testing.T) {
	convey.Convey("SetArticleUpListCache", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			list = []*model.Info{}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetArticleUpListCache(c, list)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
