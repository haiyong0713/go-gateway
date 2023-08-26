package bnj

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSendLiveItem(t *testing.T) {
	convey.Convey("SendLiveItem", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(2080809)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SendLiveItem(c, mid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
