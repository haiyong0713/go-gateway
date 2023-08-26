package prediction

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestPredictionKey(t *testing.T) {
	convey.Convey("predictionKey", t, func(ctx convey.C) {
		var (
			id = int64(7)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := predictionKey(id)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestPredItemKey(t *testing.T) {
	convey.Convey("predItemKey", t, func(ctx convey.C) {
		var (
			id = int64(7)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := predItemKey(id)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}
