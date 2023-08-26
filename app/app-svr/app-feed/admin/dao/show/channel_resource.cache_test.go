package show

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCacheAIChannelRes(t *testing.T) {
	convey.Convey("PopCRFindByTEID", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(2)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheAIChannelRes(c, id)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
