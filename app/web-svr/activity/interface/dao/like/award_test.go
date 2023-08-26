package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawAwardSubject(t *testing.T) {
	convey.Convey("RawAwardSubject", t, func(ctx convey.C) {
		var (
			c         = context.Background()
			sid int64 = 4
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RawAwardSubject(c, sid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
