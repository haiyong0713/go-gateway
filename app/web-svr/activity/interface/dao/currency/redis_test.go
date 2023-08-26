package currency

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/activity/interface/model/currency"

	"github.com/smartystreets/goconvey/convey"
)

func TestCurrencylockNumKey(t *testing.T) {
	convey.Convey("lockNumKey", t, func(convCtx convey.C) {
		var (
			sid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := lockNumKey(sid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencyunlockStateKey(t *testing.T) {
	convey.Convey("unlockStateKey", t, func(convCtx convey.C) {
		var (
			date = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := unlockStateKey(date)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencymikuStateKey(t *testing.T) {
	convey.Convey("mikuStateKey", t, func(convCtx convey.C) {
		var (
			sid = int64(1)
			mid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := singleStateKey(sid, mid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencyCacheLockNum(t *testing.T) {
	convey.Convey("CacheLockNum", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLockNum(c, sid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencyAddCacheLockNum(t *testing.T) {
	convey.Convey("AddCacheLockNum", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			count, err := d.AddCacheLockNum(c, sid)
			convCtx.Convey("Then err should be nil.count should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(count, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencyCacheUnlockState(t *testing.T) {
	convey.Convey("CacheUnlockState", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			date = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUnlockState(c, date)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestCurrencyAddCacheUnlockState(t *testing.T) {
	convey.Convey("AddCacheUnlockState", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			date = ""
			val  = int(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.AddCacheUnlockState(c, date, val)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestCurrencyCacheMikuAward(t *testing.T) {
	convey.Convey("CacheMikuState", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1)
			mid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheMikuAward(c, sid, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestCurrencyCacheSingleAward(t *testing.T) {
	convey.Convey("CacheMikuState", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1)
			mid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheSingleAward(c, sid, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestCurrencySetCacheSingleAward(t *testing.T) {
	convey.Convey("TestCurrencySetCacheSingleAward", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(1)
			mid  = int64(1)
			num  = int(0)
			data = &currency.SingleAward{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SetCacheSingleAward(c, sid, mid, num, data)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCurrencyCacheLikeTotal(t *testing.T) {
	convey.Convey("TestCurrencyCacheLikeTotal", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheLikeTotal(c, fmt.Sprintf("scholarship_test_%d", mid), mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}
