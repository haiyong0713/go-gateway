package like

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_RawClockInSubIDs(t *testing.T) {
	convey.Convey("RawClockInSubIDs", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			now = time.Now()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			ids, err := d.RawClockInSubIDs(c, now)
			ctx.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", ids)
			})
		})
	})
}
