package lottery

import (
	"context"
	lottery "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawLottery(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RawLottery", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawLottery(c, "1")
			convCtx.Convey("Then err should nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawLotteryInfo(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RawLotteryInfo", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawLotteryInfo(c, "1")
			convCtx.Convey("Then err should nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawLotteryTimesConfig(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RawLotteryTimesConfig", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawLotteryTimesConfig(c, "1")
			convCtx.Convey("Then err should nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawLotteryGift(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RawLotteryGift", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawLotteryGift(c, "1")
			convCtx.Convey("Then err should nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawLotteryAddrCheck(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RawLotteryAddrCheck", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawLotteryAddrCheck(c, 1, 1)
			convCtx.Convey("Then err should nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestInsertLotteryAddr(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("InsertLotteryAddr", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.InsertLotteryAddr(c, 1, 1, 1)
			convCtx.Convey("Then err should nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestInsertLotteryAddTimes(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("InsertLotteryAddTimes", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.InsertLotteryAddTimes(c, int64(1), int64(1), int(1), int(1), int64(1), "1", "1")
			convCtx.Convey("Then err should nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestInsertLotteryRecard(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("InsertLotteryRecard", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.InsertLotteryRecard(c, int64(1), []*lottery.InsertRecord{{Mid: 1, Num: 1, Type: 1, CID: 1}}, []int64{1}, "1")
			convCtx.Convey("Then err should nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
