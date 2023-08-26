package like

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeRawLotteryLikeList(t *testing.T) {
	convey.Convey("RawLotteryLikeList", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawLotteryLikeList(c, 5)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeCacheLotteryList(t *testing.T) {
	convey.Convey("CacheLotteryList", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLotteryList(c, 5)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestLikeAddCacheLotteryList(t *testing.T) {
	convey.Convey("AddCacheLotteryList", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			list = make(map[string]*like.Lottery)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryList(c, list, 5)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeLotteryLikeList(t *testing.T) {
	convey.Convey("LotteryLikeList", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryLikeList(c, 5)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeGoAddLotteryTimes(t *testing.T) {
	convey.Convey("GoAddLotteryTimes", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			sid        = "cc5d5ca1-0f5e-11ea-bfa0-246e9693a590"
			mid        = int64(27515241)
			actionType = int(5)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.GoAddLotteryTimes(c, sid, 98, mid, actionType, "")
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
