package bnj

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestBnjLiveGift(t *testing.T) {
	convey.Convey("LiveGift", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			mid      = int64(2089809)
			roomID   = int64(0)
			indexes  = []string{"log_user_action_102_2019_01_23", "log_user_action_102_2019_01_24"}
			timeFrom = time.Unix(1547913600, 0)
			timeTo   = time.Unix(1548086400, 0)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			result, err := d.LiveGift(c, mid, roomID, indexes, timeFrom, timeTo)
			ctx.Convey("Then err should be nil.result should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(result, convey.ShouldNotBeNil)
			})
		})
	})
}
