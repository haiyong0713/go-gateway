package lottery

import (
	"context"
	"testing"

	l "go-gateway/app/web-svr/activity/interface/model/lottery"

	"github.com/smartystreets/goconvey/convey"
)

func TestLotteryactLotteryKey(t *testing.T) {
	convey.Convey("actLotteryKey", t, func(convCtx convey.C) {
		var (
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := actLotteryKey(sid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryactLotteryInfoKey(t *testing.T) {
	convey.Convey("actLotteryInfoKey", t, func(convCtx convey.C) {
		var (
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := actLotteryInfoKey(sid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryactLotteryTimesConfKey(t *testing.T) {
	convey.Convey("actLotteryTimesConfKey", t, func(convCtx convey.C) {
		var (
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := actLotteryTimesConfKey(sid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryactLotteryGiftKey(t *testing.T) {
	convey.Convey("actLotteryGiftKey", t, func(convCtx convey.C) {
		var (
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := actLotteryGiftKey(sid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotterylotteryAddrCheckKey(t *testing.T) {
	convey.Convey("lotteryAddrCheckKey", t, func(convCtx convey.C) {
		var (
			mid = int64(88889062)
			id  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := lotteryAddrCheckKey(id, mid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryipReqKey(t *testing.T) {
	convey.Convey("ipReqKey", t, func(convCtx convey.C) {
		var (
			ip = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := ipReqKey(ip)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotterylotteryTimesKey(t *testing.T) {
	convey.Convey("lotteryTimesKey", t, func(convCtx convey.C) {
		var (
			sid    = int64(1)
			mid    = int64(1)
			remark = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := lotteryTimesKey(sid, mid, remark)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotterylotteryTimesField(t *testing.T) {
	convey.Convey("lotteryTimesKey", t, func(convCtx convey.C) {
		var (
			ltc = &l.LotteryTimesConfig{ID: 1, Type: 1, AddType: 1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := lotteryTimesField(ltc)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotterylotteryWinListKey(t *testing.T) {
	convey.Convey("lotteryWinListKey", t, func(convCtx convey.C) {
		var (
			sid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := lotteryWinListKey(sid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotterylotteryActionKey(t *testing.T) {
	convey.Convey("lotteryActionKey", t, func(convCtx convey.C) {
		var (
			sid = int64(1)
			mid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := lotteryActionKey(sid, mid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotterylotteryGiftNumKey(t *testing.T) {
	convey.Convey("lotteryGiftNumKey", t, func(convCtx convey.C) {
		var (
			sid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := lotteryGiftNumKey(sid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotterylotteryTimesMapKey(t *testing.T) {
	convey.Convey("lotteryGiftNumKey", t, func(convCtx convey.C) {
		var (
			val = &l.LotteryTimesConfig{ID: 1, Type: 1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := lotteryTimesMapKey(val)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryCacheLottery(t *testing.T) {
	convey.Convey("CacheLottery", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLottery(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheLottery(t *testing.T) {
	convey.Convey("AddCacheLottery", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
			val = &l.Lottery{ID: 1, LotteryID: "asdfghjk", Name: "test"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLottery(c, sid, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryDeleteLottery(t *testing.T) {
	convey.Convey("DeleteLottery", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DeleteLottery(c, sid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryCacheLotteryInfo(t *testing.T) {
	convey.Convey("CacheLotteryInfo", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLotteryInfo(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheLotteryInfo(t *testing.T) {
	convey.Convey("AddCacheLotteryInfo", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
			val = &l.LotteryInfo{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryInfo(c, sid, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryDeleteLotteryInfo(t *testing.T) {
	convey.Convey("DeleteLotteryInfo", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DeleteLotteryInfo(c, sid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryCacheLotteryTimesConfig(t *testing.T) {
	convey.Convey("CacheLotteryTimesConfig", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLotteryTimesConfig(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheLotteryTimesConfig(t *testing.T) {
	convey.Convey("AddCacheLotteryTimesConfig", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = "asdfghjk"
			list = []*l.LotteryTimesConfig{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryTimesConfig(c, sid, list)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryDeleteLotteryTimesConfig(t *testing.T) {
	convey.Convey("DeleteLotteryTimesConfig", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DeleteLotteryTimesConfig(c, sid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryCacheLotteryGift(t *testing.T) {
	convey.Convey("CacheLotteryGift", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLotteryGift(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheLotteryGift(t *testing.T) {
	convey.Convey("AddCacheLotteryGift", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = "asdfghjk"
			list = []*l.LotteryGift{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryGift(c, sid, list)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryDeleteLotteryGift(t *testing.T) {
	convey.Convey("DeleteLotteryGift", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = "asdfghjk"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DeleteLotteryGift(c, sid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryCacheLotteryAddrCheck(t *testing.T) {
	convey.Convey("CacheLotteryAddrCheck", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(0)
			id  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLotteryAddrCheck(c, id, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheLotteryAddrCheck(t *testing.T) {
	convey.Convey("AddCacheLotteryAddrCheck", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(0)
			id  = int64(1)
			val = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryAddrCheck(c, id, mid, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryCacheIPRequestCheck(t *testing.T) {
	convey.Convey("CacheIPRequestCheck", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			ip = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheIPRequestCheck(c, ip)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheIPRequestCheck(t *testing.T) {
	convey.Convey("AddCacheIPRequestCheck", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ip  = ""
			val = int(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheIPRequestCheck(c, ip, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryCacheLotteryTimes(t *testing.T) {
	convey.Convey("CacheLotteryTimes", t, func(convCtx convey.C) {
		var (
			c                  = context.Background()
			sid                = int64(1)
			mid                = int64(0)
			lotteryTimesConfig = []*l.LotteryTimesConfig{}
			remark             = "add"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CacheLotteryTimes(c, sid, mid, lotteryTimesConfig, remark)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestLotteryAddCacheLotteryTimes(t *testing.T) {
	convey.Convey("AddCacheLotteryTimes", t, func(convCtx convey.C) {
		var (
			c                  = context.Background()
			sid                = int64(1)
			mid                = int64(1)
			lotteryTimesConfig = []*l.LotteryTimesConfig{{ID: 1, Type: 1, AddType: 1}, {ID: 2, Type: 3, AddType: 1}}
			remark             = "add"
			list               = make(map[string]int)
		)
		list["190"] = 1
		list["291"] = 1
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryTimes(c, sid, mid, lotteryTimesConfig, remark, list)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryIncrTimes(t *testing.T) {
	convey.Convey("IncrTimes", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			sid    = int64(1)
			mid    = int64(0)
			ltc    = &l.LotteryTimesConfig{}
			val    = int(1)
			status = "add"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.IncrTimes(c, sid, mid, ltc, val, status)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryCacheLotteryWinList(t *testing.T) {
	convey.Convey("CacheLotteryWinList", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLotteryWinList(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheLotteryWinList(t *testing.T) {
	convey.Convey("AddCacheLotteryWinList", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(1)
			list = []*l.GiftList{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryWinList(c, sid, list)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryCacheLotteryActionLog(t *testing.T) {
	convey.Convey("CacheLotteryActionLog", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(1)
			mid   = int64(0)
			start = int64(0)
			end   = int64(5)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLotteryActionLog(c, sid, mid, start, end)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryAddCacheLotteryActionLog(t *testing.T) {
	convey.Convey("AddCacheLotteryActionLog", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(1)
			mid  = int64(0)
			lr   = &l.LotteryRecordDetail{Mid: 0, Num: 1, GiftID: 0, Type: 0, CID: 0}
			list = []*l.LotteryRecordDetail{}
		)
		list = append(list, lr)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheLotteryActionLog(c, sid, mid, list)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryAddLotteryActionLog(t *testing.T) {
	convey.Convey("AddLotteryActionLog", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(1)
			mid  = int64(0)
			lr   = &l.LotteryRecordDetail{Mid: 0, Num: 1, GiftID: 0, Type: 0, CID: 0}
			list = []*l.LotteryRecordDetail{}
		)
		list = append(list, lr)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddLotteryActionLog(c, sid, mid, list)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLotteryCacheGiftNum(t *testing.T) {
	convey.Convey("CacheGiftNum", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			sid     = int64(1)
			giftMap = make(map[int64]*l.LotteryGift)
		)
		lg := &l.LotteryGift{ID: 1, Sid: "asdfghjk"}
		giftMap[1] = lg
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			num, err := d.CacheGiftNum(c, sid, giftMap)
			convCtx.Convey("Then err should be nil.num should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(num, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryIncrGiftNum(t *testing.T) {
	convey.Convey("IncrGiftNum", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			sid    = int64(1)
			giftID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.IncrGiftNum(c, sid, giftID)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
