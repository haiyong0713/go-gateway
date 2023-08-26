package article

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestUpSwitch(t *testing.T) {
	c := context.Background()
	convey.Convey("UpSwitch", t, func(ctx convey.C) {
		res, err := d.UpSwitch(c, 1111112639)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldBeTrue)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAccCards(t *testing.T) {
	c := context.Background()
	convey.Convey("AccCards", t, func(ctx convey.C) {
		res, err := d.AccCards(c, []int64{9999999999999, 9999999999})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
