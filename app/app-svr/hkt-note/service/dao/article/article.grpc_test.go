package article

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestArticleAudits(t *testing.T) {
	c := context.Background()
	convey.Convey("ArticleAudits", t, func(ctx convey.C) {
		res, err := d.ArticleAudits(c, []int64{4543})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}
