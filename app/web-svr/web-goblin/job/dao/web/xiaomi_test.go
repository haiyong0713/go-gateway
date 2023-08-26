package web

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_DelXiaomiArc(t *testing.T) {
	var (
		bvid        = "BV1JE411V765"
		total int64 = 1
		c           = context.Background()
	)
	convey.Convey("OutArcByMtime", t, func(ctx convey.C) {
		err := d.DelXiaomiArc(c, total, bvid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
