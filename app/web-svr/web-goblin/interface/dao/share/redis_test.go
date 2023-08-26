package share

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSharesCache(t *testing.T) {
	convey.Convey("SharesCache", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SharesCache(context.Background(), 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestSetSharesCache(t *testing.T) {
	convey.Convey("SetSharesCache", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetSharesCache(context.Background(), 1000, 1, map[string]int64{"aid": 1})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
