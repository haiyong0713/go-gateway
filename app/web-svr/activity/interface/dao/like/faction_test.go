package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_FactionRank(t *testing.T) {
	convey.Convey("ChildhoodRank", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FactionRank(c)
			ctx.Convey("Then err should be nil.likes should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%v", res)
			})
		})
	})
}
