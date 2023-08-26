package show

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawAggregation(t *testing.T) {
	convey.Convey("TestRawAggregation", t, func(convCtx convey.C) {
		var (
			ctx = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := dao.RawAggregation(ctx, 1)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestRawAggregations(t *testing.T) {
	convey.Convey("TestRawAggregations", t, func(convCtx convey.C) {
		var (
			ctx   = context.Background()
			hotID = []int64{1, 2, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := dao.RawAggregations(ctx, hotID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestAIAggregation(t *testing.T) {
	convey.Convey("TestAIAggregation", t, func(convCtx convey.C) {
		var (
			ctx = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := dao.AIAggregation(ctx, 1)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				println(err)
			})
		})
	})
}
