package bws

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawActFields(t *testing.T) {
	convey.Convey("RawActFields", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawActFields(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}
