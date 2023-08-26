package prediction

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestPredictions(t *testing.T) {
	convey.Convey("Predictions", t, func(ctx convey.C) {
		var (
			ids = []int64{1, 2, 3}
			c   = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			list, err := d.Predictions(c, ids)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Print(list)
			})
		})
	})
}

func TestPredItems(t *testing.T) {
	convey.Convey("PredItems", t, func(ctx convey.C) {
		var (
			ids = []int64{1, 2, 3}
			c   = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			list, err := d.PredItems(c, ids)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Print(list)
			})
		})
	})
}
