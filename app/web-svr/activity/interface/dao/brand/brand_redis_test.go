package brand

import (
	"context"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestCacheAddCouponTimes(t *testing.T) {
	convey.Convey("CacheAddCouponTimes", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheAddCouponTimes(c, 1111111)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("when eveything goes positive too", func(convCtx convey.C) {
			_, err := d.CacheAddCouponTimes(c, 1112111)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheQPSLimit(t *testing.T) {
	convey.Convey("TestCacheQPSLimit", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheQPSLimit(c, "counpon")
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("when eveything goes positive too", func(convCtx convey.C) {
			_, err := d.CacheQPSLimit(c, "counpon")
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
