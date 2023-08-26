package lottery

import (
	"context"
	"testing"
	"time"

	l "go-gateway/app/web-svr/activity/interface/model/lottery"
	lotterymdl "go-gateway/app/web-svr/activity/interface/model/lottery"

	"github.com/smartystreets/goconvey/convey"
)

func TestLotterylotteryMcNumKey(t *testing.T) {
	convey.Convey("lotteryMcNumKey", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res := lotteryMcNumKey(1, 1, 1)
			convCtx.Convey("Then res should not be nil.", func(convCtx convey.C) {
				convCtx.Print(res)
			})
		})
	})
}

func TestLotteryrsKey(t *testing.T) {
	convey.Convey("rsKey", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res := rsKey(1, 1)
			convCtx.Convey("Then res should not be nil.", func(convCtx convey.C) {
				convCtx.Print(res)
			})
		})
	})
}

func TestLotteryRawLottery(t *testing.T) {
	convey.Convey("RawLottery", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjkl"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLottery(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryInfo(t *testing.T) {
	convey.Convey("RawLotteryInfo", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjkl"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryInfo(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryTimesConfig(t *testing.T) {
	convey.Convey("RawLotteryTimesConfig", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjkl"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryTimesConfig(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryGift(t *testing.T) {
	convey.Convey("RawLotteryGift", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjkl"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryGift(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryAddrCheck(t *testing.T) {
	convey.Convey("RawLotteryAddrCheck", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(0)
			id  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryAddrCheck(c, id, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryInsertLotteryAddr(t *testing.T) {
	convey.Convey("InsertLotteryAddr", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			mid       = int64(0)
			addressId = int64(0)
			id        = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			ef, err := d.InsertLotteryAddr(c, id, mid, addressId)
			convCtx.Convey("Then err should be nil.ef should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(ef, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryInsertLotteryAddTimes(t *testing.T) {
	convey.Convey("InsertLotteryAddTimes", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			id      = int64(21)
			mid     = int64(12540164)
			addType = int(3)
			num     = int(1)
			cid     = int64(3)
			ip      = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			ef, err := d.InsertLotteryAddTimes(c, id, mid, addType, num, cid, ip, "")
			convCtx.Convey("Then err should be nil.ef should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(ef, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryInsertLotteryRecard(t *testing.T) {
	convey.Convey("InsertLotteryRecard", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			id     = int64(1)
			record = []*lotterymdl.InsertRecord{}
			gid    = []int64{}
			ip     = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			count, err := d.InsertLotteryRecard(c, id, record, gid, ip)
			convCtx.Convey("Then err should be nil.count should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(count, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryAddTimes(t *testing.T) {
	convey.Convey("RawLotteryAddTimes", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(1)
			mid = int64(27515241)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryAddTimes(c, id, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryUsedTimes(t *testing.T) {
	convey.Convey("RawLotteryUsedTimes", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(17)
			mid = int64(27515241)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryUsedTimes(c, id, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryGetMemberAddress(t *testing.T) {
	convey.Convey("GetMemberAddress", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(1002)
			mid = int64(88889062)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			val, err := d.GetMemberAddress(c, id, mid)
			convCtx.Convey("Then err should be nil.val should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(val, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryWinList(t *testing.T) {
	convey.Convey("RawLotteryWinList", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(1)
			num = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryWinList(c, id, []int64{}, num)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryInsertLotteryWin(t *testing.T) {
	convey.Convey("InsertLotteryWin", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			id     = int64(1)
			giftID = int64(0)
			mid    = int64(0)
			ip     = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			ef, err := d.InsertLotteryWin(c, id, giftID, mid, ip)
			convCtx.Convey("Then err should be nil.ef should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(ef, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryUpdateLotteryWin(t *testing.T) {
	convey.Convey("UpdateLotteryWin", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			id     = int64(1)
			mid    = int64(123456)
			giftID = int64(1)
			ip     = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			ef, err := d.UpdateLotteryWin(c, id, mid, giftID, ip)
			convCtx.Convey("Then err should be nil.ef should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(ef, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryWinOne(t *testing.T) {
	convey.Convey("RawLotteryWinOne", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			id     = int64(1)
			mid    = int64(123456)
			giftID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryWinOne(c, id, mid, giftID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLottery(t *testing.T) {
	convey.Convey("Lottery", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjkl"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Lottery(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLotteryInfo(t *testing.T) {
	convey.Convey("LotteryInfo", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjkl"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryInfo(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLotteryTimesConfig(t *testing.T) {
	convey.Convey("LotteryTimesConfig", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjkl"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryTimesConfig(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLotteryGift(t *testing.T) {
	convey.Convey("LotteryGift", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjkl"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryGift(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLotteryAddr(t *testing.T) {
	convey.Convey("LotteryAddr", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(0)
			id  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryAddr(c, id, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLotteryUsedTimes(t *testing.T) {
	convey.Convey("LotteryUsedTimes", t, func(convCtx convey.C) {
		var (
			c                  = context.Background()
			lotteryTimesConfig = []*l.LotteryTimesConfig{}
			id                 = int64(17)
			mid                = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryUsedTimes(c, lotteryTimesConfig, id, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLotteryAddTimes(t *testing.T) {
	convey.Convey("LotteryAddTimes", t, func(convCtx convey.C) {
		var (
			c                  = context.Background()
			lotteryTimesConfig = []*l.LotteryTimesConfig{}
			id                 = int64(1)
			mid                = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryAddTimes(c, lotteryTimesConfig, id, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLotteryWinList(t *testing.T) {
	convey.Convey("LotteryWinList", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(1)
			num = int64(5)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryWinList(c, id, []int64{}, num, false)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryCacheLotteryMcNum(t *testing.T) {
	convey.Convey("CacheLotteryMcNum", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(1)
			high = int(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLotteryMcNum(c, sid, high, 0)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheLotteryMcNum(t *testing.T) {
	convey.Convey("AddCacheLotteryMcNum", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(1)
			high = int(0)
			val  = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryMcNum(c, sid, high, 0, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotterySendSysMsg(t *testing.T) {
	convey.Convey("SendSysMsg", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			uids    = []int64{27515241}
			mc      = "1_4_1"
			title   = "test"
			context = "test"
			ip      = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SendSysMsg(c, uids, mc, title, context, ip)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryMemberCoupon(t *testing.T) {
	convey.Convey("MemberCoupon", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			mid        = 27515241
			batchToken = "841764801720181030122848"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.MemberCoupon(c, int64(mid), batchToken)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryMemberVip(t *testing.T) {
	convey.Convey("MemberVip", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			mid        = 27515241
			batchToken = "test"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.MemberVip(c, int64(mid), batchToken, "test")
			err = nil
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryLotteryGiftNum(t *testing.T) {
	convey.Convey("RawLotteryUsedTimes", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sids = []string{"asdfghjkl", "qwertyuiop"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotteryGiftNum(c, sids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryLotterySid(t *testing.T) {
	convey.Convey("RawLotteryUsedTimes", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			nowT = time.Now().Unix()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.LotterySid(c, nowT)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryRawLotteryOrderNo(t *testing.T) {
	convey.Convey("RawLotteryOrderNo", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			id      = int64(21)
			mid     = int64(110)
			orderNo = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawLotteryOrderNo(c, id, mid, orderNo)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}
