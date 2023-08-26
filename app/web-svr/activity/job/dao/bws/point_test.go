package bws

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestBwsbwsLotteryKey(t *testing.T) {
	convey.Convey("bwsLotteryKey", t, func(convCtx convey.C) {
		var (
			bid     = int64(0)
			awardID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := bwsLotteryKey(bid, awardID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsRechargeAward(t *testing.T) {
	convey.Convey("RechargeAward", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			awards, err := d.RechargeAward(c, bid)
			convCtx.Convey("Then err should be nil.awards should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(awards, convey.ShouldNotBeNil)
				bs, _ := json.Marshal(awards)
				convCtx.Println(string(bs))
			})
		})
	})
}
