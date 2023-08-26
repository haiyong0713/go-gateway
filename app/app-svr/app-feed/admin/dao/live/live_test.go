package live

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestLiveAppMRoom(t *testing.T) {
	convey.Convey("LiveRoom", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			roomids = []int64{12133}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			rs, err := d.LiveRoom(c, roomids)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
		})
	})
}
