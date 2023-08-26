package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoArticleGiant(t *testing.T) {
	convey.Convey("Article giant", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(27515401)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.ArticleGiant(c, mid)
			ctx.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestDaoArticleGiantV4Reset(t *testing.T) {
	convey.Convey("Article giant v4", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(88895249)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.ArticleGiantV4Reset(c, mid)
			ctx.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoArticleGiantV4(t *testing.T) {
	convey.Convey("Article giant v4", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(88895249)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.ArticleGiantV4(c, mid)
			ctx.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestDaoArticleLists(t *testing.T) {
	convey.Convey("ArticleLists", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1729}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.ArticleLists(c, ids)
			ctx.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestDaoUpArtLists(t *testing.T) {
	convey.Convey("UpArtLists", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(27515401)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.UpArtLists(c, mid)
			ctx.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}
