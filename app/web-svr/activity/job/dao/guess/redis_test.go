package guess

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_DelOidGuessCache(t *testing.T) {
	convey.Convey("DelOidGuessCache", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			oid      = int64(1)
			business = int64(1)
			mainID   = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelGuessCache(c, oid, business, mainID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_DelUserStat(t *testing.T) {
	convey.Convey("DelUserStat", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			mid       = int64(1)
			stakeType = int64(1)
			business  = int64(1)
			mainID    = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelUserCache(c, mid, stakeType, business, mainID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_DelStatCache(t *testing.T) {
	convey.Convey("DelStatCache", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			mid       = int64(1)
			stakeType = int64(1)
			business  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelStatCache(c, mid, stakeType, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
