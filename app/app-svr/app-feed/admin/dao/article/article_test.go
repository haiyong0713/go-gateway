package article

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestArticleArticle(t *testing.T) {
	convey.Convey("Article", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			articleids = []int64{568}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.Article(c, articleids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestArticleTitles(t *testing.T) {
	convey.Convey("Article", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			articleids = []int64{568, 2133}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.ArticlesInfo(c, articleids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func Test_ArticleRpc(t *testing.T) {
	convey.Convey("ArticleRpc", t, func(ctx convey.C) {
		var (
			c               = context.Background()
			articleId int64 = 2133
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.ArticleRpc(c, articleId)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
