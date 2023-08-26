package comic

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestComicCoupon(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(111)
	)
	convey.Convey("Coupon", t, func(ctx convey.C) {
		httpMock("POST", d.comicCouponURL).Reply(200).JSON(`{"code":0,"msg":"succ"}`)
		msg, err := d.Coupon(c, mid)
		ctx.Convey("Then err should be nil.msg should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(msg, convey.ShouldNotBeNil)
			fmt.Println(msg)
		})
	})
}

func TestComicComicUser(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(111)
	)
	convey.Convey("ComicUser", t, func(ctx convey.C) {
		httpMock("POST", d.comicUserURL).Reply(200).JSON(`{"code":0,"msg":"succ","data":true}`)
		isUser, err := d.ComicUser(c, mid)
		ctx.Convey("Then err should be nil.msg should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(isUser, convey.ShouldNotBeNil)
			fmt.Println(isUser)
		})
	})
}
