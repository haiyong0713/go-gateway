package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoWallet(t *testing.T) {
	convey.Convey("Wallet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			wallet, err := d.Wallet(c, mid)
			ctx.Convey("Then err should be nil.wallet should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(wallet, convey.ShouldNotBeNil)
			})
		})
	})
}
