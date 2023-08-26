package web

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestOutArcByMtime(t *testing.T) {
	var (
		c       = context.Background()
		nowTime = time.Now()
		from    = nowTime.AddDate(0, 0, -1)
		to      = nowTime.AddDate(0, 0, -1)
	)
	convey.Convey("OutArcByMtime", t, func(ctx convey.C) {
		data, err := d.OutArcByMtime(c, from, to)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v", data)
		})
	})
}
