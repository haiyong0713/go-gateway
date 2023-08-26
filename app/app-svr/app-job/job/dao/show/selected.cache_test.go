package show

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCardPickSerieCache(t *testing.T) {
	var (
		c      = context.Background()
		sType  = "weekly_selected"
		number = int64(1)
	)
	convey.Convey("PickSerieCache", t, func(ctx convey.C) {
		serie, err := d.PickSerieCache(c, sType, number)
		ctx.Convey("Then err should be nil.serie should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(serie, convey.ShouldNotBeNil)
		})
	})
}
