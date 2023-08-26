package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeSteinRuleCount(t *testing.T) {
	convey.Convey("SteinRuleCount", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			sid      = int64(10541)
			viewRule = int64(1)
			likeRule = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			total, err := d.SteinRuleCount(c, sid, viewRule, likeRule)
			ctx.Convey("Then err should be nil.total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
				ctx.Printf("%+v", total)
			})
		})
	})
}
