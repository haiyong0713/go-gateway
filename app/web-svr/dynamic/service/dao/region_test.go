package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoArchiveAll(t *testing.T) {
	convey.Convey("ArchiveAll", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			rid    = int32(32)
			start  = int(1)
			length = int(50)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.ArchiveAll(c, rid, start, length)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
