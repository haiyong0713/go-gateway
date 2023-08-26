package article

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDelUpArticles(t *testing.T) {
	c := context.Background()
	convey.Convey("DelUpArticles", t, func(ctx convey.C) {
		err := d.DelUpArticles(c, []int64{4425, 4434, 4433, 4432}, 27515244)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
