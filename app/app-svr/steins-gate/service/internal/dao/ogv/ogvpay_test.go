package ogv

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoOgvPay(t *testing.T) {
	convey.Convey("OgvPay", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
			aid = int64(23)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			allowPlay, err := d.OgvPay(c, mid, aid)
			fmt.Println(allowPlay, err)
			convCtx.Convey("Then err should be nil.allowPlay should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(allowPlay, convey.ShouldNotBeNil)
			})
		})
	})
}
