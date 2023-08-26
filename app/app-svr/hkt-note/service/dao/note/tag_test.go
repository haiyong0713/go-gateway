package note

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestToTags(t *testing.T) {
	convey.Convey("ToTags", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			aid    = int64(720019004)
			noteID = int64(873056727728142)
			tagStr = "10217680-1-0-0,10217680-1-19-1,10217680-1-0-2,10217680-1-0-3"
		)
		res, _, err := d.ToTags(c, aid, noteID, tagStr, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}
