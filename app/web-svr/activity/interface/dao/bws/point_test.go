package bws

import (
	"context"
	"testing"

	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/smartystreets/goconvey/convey"
)

func TestBwskeyUserPoint(t *testing.T) {
	convey.Convey("keyUserPoint", t, func(convCtx convey.C) {
		var (
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := keyUserPoint(bid, key)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsRawPoints(t *testing.T) {
	convey.Convey("RawPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawPoints(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestRawBwsPoints(t *testing.T) {
	convey.Convey("RawBwsPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = []int64{1, 7}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawBwsPoints(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsRawUserPoints(t *testing.T) {
	convey.Convey("RawUserPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			rs, err := d.RawUserPoints(c, bid, key)
			convCtx.Convey("Then err should be nil.rs should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", rs)
			})
		})
	})
}

func TestBwsCacheUserPoints(t *testing.T) {
	convey.Convey("CacheUserPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUserPoints(c, bid, key)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", res)
			})
		})
	})
}

func TestBwsAddCacheUserPoints(t *testing.T) {
	convey.Convey("AddCacheUserPoints", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			bid  = int64(0)
			data = []*bwsmdl.UserPoint{}
			key  = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserPoints(c, bid, data, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsAppendUserPointsCache(t *testing.T) {
	convey.Convey("AppendUserPointsCache", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			bid   = int64(0)
			key   = ""
			point = &bwsmdl.UserPoint{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AppendUserPointsCache(c, bid, key, point)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheUserPoints(t *testing.T) {
	convey.Convey("DelCacheUserPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheUserPoints(c, bid, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestIncrUnlock(t *testing.T) {
	convey.Convey("IncrUnlock", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			pid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.IncrUnlock(c, pid, 1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddUnlock(t *testing.T) {
	convey.Convey("AddUnlock", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			pid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddUnlock(c, pid, 1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// CacheUnlock
func TestCacheUnlock(t *testing.T) {
	convey.Convey("CacheUnlock", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			pid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheUnlock(c, pid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawCountUnlock(t *testing.T) {
	convey.Convey("RawCountUnlock", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			pid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawCountUnlock(c, pid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawUserLockPoints(t *testing.T) {
	convey.Convey("RawUserLockPoints", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			bid      = int64(3)
			lockType = int64(2)
			key      = "9875fa517967622b"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawUserLockPoints(c, bid, lockType, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestRawUserHp(t *testing.T) {
	convey.Convey("RawUserHp", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(3)
			key = "9875fa517967622b"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawUserHp(c, bid, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestAddUserPoint(t *testing.T) {
	convey.Convey("AddUserPoint", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(2)
			pid = int64(5)
			key = "9875fa517967622b"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.AddUserPoint(c, bid, pid, 4, -1, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestRechargePoint(t *testing.T) {
	convey.Convey("RechargePoint", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			pid = int64(4)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.RechargePoint(c, pid, 2, -1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsAddCacheUserLockPoints(t *testing.T) {
	convey.Convey("AddCacheUserLockPoints", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			bid      = int64(2)
			data     = []*bwsmdl.UserPoint{{ID: 1, Pid: 2, Points: 2, Ctime: 1562035067}}
			key      = "123456"
			lockType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserLockPoints(c, bid, data, lockType, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAppendUserLockPointsCache(t *testing.T) {
	convey.Convey("AppendUserLockPointsCache", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			bid      = int64(2)
			data     = &bwsmdl.UserPoint{ID: 1, Pid: 2, Points: 2, Ctime: 1562035067}
			key      = "123456"
			lockType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AppendUserLockPointsCache(c, bid, lockType, key, data)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheUserLockPoints(t *testing.T) {
	convey.Convey("CacheUserLockPoints", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			bid      = int64(2)
			key      = "123456"
			lockType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CacheUserLockPoints(c, bid, lockType, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestDelCacheUserLockPoints(t *testing.T) {
	convey.Convey("DelCacheUserLockPoints", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			bid      = int64(2)
			key      = "123456"
			lockType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheUserLockPoints(c, bid, lockType, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheUserHp(t *testing.T) {
	convey.Convey("CacheUserHp", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(2)
			key = "123456"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUserHp(c, bid, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestAddCacheUserHp(t *testing.T) {
	convey.Convey("AddCacheUserHp", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(2)
			key = "123456"
			val = int64(-1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserHp(c, bid, val, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestIncrUserHp(t *testing.T) {
	convey.Convey("IncrUserHp", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(2)
			key = "123456"
			val = int64(5)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.IncrUserHp(c, bid, val, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPointsByBid(t *testing.T) {
	convey.Convey("PointsByBid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.PointsByBid(c, bid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

// DelCacheUserHp
func TestDelCacheUserHp(t *testing.T) {
	convey.Convey("DelCacheUserHp", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(2)
			key = "1234"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheUserHp(c, bid, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// RawPointsByLock
func TestRawPointsByLock(t *testing.T) {
	convey.Convey("RawPointsByLock", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawPointsByLock(c, bid, 6)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", res)
			})
		})
	})
}

func TestBatchUserLockPoints(t *testing.T) {
	convey.Convey("BatchUserLockPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.BatchUserLockPoints(c, bid, []int64{5, 6}, "1234")
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", res)
			})
		})
	})
}

func TestCacheBatchUserLockPoints(t *testing.T) {
	convey.Convey("CacheBatchUserLockPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheBatchUserLockPoints(c, bid, []int64{5, 6}, "1234")
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", res)
			})
		})
	})
}

func TestRawBatchUserLockPoints(t *testing.T) {
	convey.Convey("RawBatchUserLockPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawBatchUserLockPoints(c, bid, []int64{5, 6}, "1234")
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", res)
			})
		})
	})
}

func TestAddCacheBatchUserLockPoints(t *testing.T) {
	convey.Convey("AddCacheBatchUserLockPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
			val = make(map[int64][]*bwsmdl.UserPoint)
		)
		val[5] = append(val[5], &bwsmdl.UserPoint{ID: 1})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheBatchUserLockPoints(c, bid, val, "1234")
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
