package bws

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go-common/library/cache/redis"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/smartystreets/goconvey/convey"
)

func TestBwskeyUserAchieve(t *testing.T) {
	convey.Convey("keyUserAchieve", t, func(convCtx convey.C) {
		var (
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := keyUserAchieve(bid, key)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwskeyAchieveCnt(t *testing.T) {
	convey.Convey("keyAchieveCnt", t, func(convCtx convey.C) {
		var (
			bid = int64(0)
			day = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := keyAchieveCnt(bid, day)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwskeyLottery(t *testing.T) {
	convey.Convey("keyLottery", t, func(convCtx convey.C) {
		var (
			aid = int64(0)
			day = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := keyLottery(aid, day)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsAward(t *testing.T) {
	convey.Convey("Award", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			key = "c0ca560a9710be37"
			aid = int64(34)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.Award(c, key, aid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsRawAchievements(t *testing.T) {
	convey.Convey("RawAchievements", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawAchievements(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsRawUserAchieves(t *testing.T) {
	convey.Convey("RawUserAchieves", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			rs, err := d.RawUserAchieves(c, bid, key)
			convCtx.Convey("Then err should be nil.rs should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", rs)
			})
		})
	})
}

func TestBwsRawAchieveCounts(t *testing.T) {
	convey.Convey("RawAchieveCounts", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			day = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawAchieveCounts(c, bid, day)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", res)
			})
		})
	})
}

func TestBwsAddCacheUserAchieves(t *testing.T) {
	convey.Convey("AddCacheUserAchieves", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			bid  = int64(0)
			data = []*bwsmdl.UserAchieve{{ID: 1, Aid: 2, Award: 1}, {ID: 2, Aid: 3, Award: 1}}
			key  = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserAchieves(c, bid, data, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsAppendUserAchievesCache(t *testing.T) {
	convey.Convey("AppendUserAchievesCache", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			bid     = int64(0)
			key     = ""
			achieve = &bwsmdl.UserAchieve{ID: 3, Aid: 4, Award: 1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AppendUserAchievesCache(c, bid, key, achieve)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheUserAchieves(t *testing.T) {
	convey.Convey("CacheUserAchieves", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUserAchieves(c, bid, key)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsDelCacheUserAchieves(t *testing.T) {
	convey.Convey("DelCacheUserAchieves", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheUserAchieves(c, bid, key)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheAchieveCounts(t *testing.T) {
	convey.Convey("CacheAchieveCounts", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(3)
			now = time.Now()
			day = now.Format("20060102")
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheAchieveCounts(c, bid, day)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", res)
			})
		})
	})
}

func TestBwsAddCacheAchieveCounts(t *testing.T) {
	convey.Convey("AddCacheAchieveCounts", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			res = []*bwsmdl.CountAchieves{}
			day = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheAchieveCounts(c, bid, res, day)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsIncrCacheAchieveCounts(t *testing.T) {
	convey.Convey("IncrCacheAchieveCounts", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			aid = int64(0)
			day = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.IncrCacheAchieveCounts(c, bid, aid, day)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheAchieveCounts(t *testing.T) {
	convey.Convey("DelCacheAchieveCounts", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			day = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheAchieveCounts(c, bid, day)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsAddLotteryMidCache(t *testing.T) {
	convey.Convey("AddLotteryMidCache", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(0)
			mid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddLotteryMidCache(c, aid, mid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheLotteryMids(t *testing.T) {
	convey.Convey("CacheLotteryMids", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(0)
			now = time.Now()
			day = now.Format("20060102")
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			mids, err := d.CacheLotteryMids(c, aid, day)
			convCtx.Convey("Then err should be nil.mids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", mids)
			})
		})
	})
}

func TestBwsCacheLotteryMid(t *testing.T) {
	convey.Convey("CacheLotteryMid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(0)
			now = time.Now()
			day = now.Format("20060102")
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			mid, err := d.CacheLotteryMid(c, aid, day)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				if err == redis.ErrNil {
					err = nil
				}
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%d", mid)
			})
		})
	})
}

func TestCacheAchievesPoint(t *testing.T) {
	convey.Convey("CacheAchievesPoint", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			bid   = int64(2)
			ukeys = []string{"123456"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheAchievesPoint(c, bid, ukeys)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddCacheAchievesPoint(t *testing.T) {
	convey.Convey("AddCacheAchievesPoint", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			bid  = int64(2)
			miss = map[string]int64{"87654": 3, "hddhhd": 4}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheAchievesPoint(c, bid, miss)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddCacheAchievesRank(t *testing.T) {
	convey.Convey("AddCacheAchievesRank", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			bid  = int64(2)
			miss = map[int64]*bwsmdl.RankAchieve{1: {Num: 5, Ctime: 1561950548}}
			ty   = 1
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheAchievesRank(c, bid, miss, ty)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheAchievesRank(t *testing.T) {
	convey.Convey("CacheAchievesRank", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
			num = 20
			ty  = 1
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CacheAchievesRank(c, bid, num, ty)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Print(list)
			})
		})
	})
}

func TestIncrAchievesPoint(t *testing.T) {
	convey.Convey("IncrAchievesPoint", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			bid   = int64(2)
			mid   = int64(17)
			score = int64(1)
			ctime = int64(1561950548)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.IncrAchievesPoint(c, bid, mid, score, ctime, false)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestIncrSingleAchievesPoint(t *testing.T) {
	convey.Convey("IncrSingleAchievesPoint", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			bid   = int64(2)
			mid   = int64(17)
			score = int64(1)
			ctime = int64(1561950548)
			ukey  = "548798678"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.IncrSingleAchievesPoint(c, bid, mid, score, ctime, ukey, false)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawCompositeAchievesPoint(t *testing.T) {
	convey.Convey("RawCompositeAchievesPoint", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = []int64{27515256}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawCompositeAchievesPoint(c, mid)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}
func TestGetAchieveRank(t *testing.T) {
	convey.Convey("GetAchieveRank", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
			mid = int64(27515256)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.GetAchieveRank(c, bid, mid, 0)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Print(list)
			})
		})
	})
}

func TestCacheCompositeAchievesPoint(t *testing.T) {
	convey.Convey("CacheCompositeAchievesPoint", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = []int64{1, 2, 3, 4}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CacheCompositeAchievesPoint(c, mid)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestAddCacheCompositeAchievesPoint(t *testing.T) {
	convey.Convey("AddCacheCompositeAchievesPoint", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = map[int64]int64{1: 3, 2: 4, 3: 6, 4: 5, 5: 15}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheCompositeAchievesPoint(c, mid)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// RawUsersAchieves
func TestRawUsersAchieves(t *testing.T) {
	convey.Convey("RawUsersAchieves", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			bid   = int64(1)
			ukesy = []string{"9abf1997abe851e6", "9abf1997abe851e5"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawUsersAchieves(c, bid, ukesy)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestRawAchievesPoint(t *testing.T) {
	convey.Convey("RawAchievesPoint", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			bid   = int64(6)
			ukesy = []string{"4457568c828b253d", "9abf1997abe851e5"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawAchievesPoint(c, bid, ukesy)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestDelCacheAchievesPoint(t *testing.T) {
	convey.Convey("RawAchievesPoint", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			bid   = int64(1)
			ukesy = []string{"1", "2", "3"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheAchievesPoint(c, bid, ukesy)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// DelCacheCompositeAchievesPoint
func TestDelCacheCompositeAchievesPoint(t *testing.T) {
	convey.Convey("DelCacheCompositeAchievesPoint", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			mids = []int64{1, 2, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheCompositeAchievesPoint(c, mids)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// LastAchievements
func TestLastAchievements(t *testing.T) {
	convey.Convey("LastAchievements", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			ukey = map[int64]string{3: "ad338a8cd0332483"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.LastAchievements(c, ukey)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestAchievesPoint(t *testing.T) {
	convey.Convey("AchievesPoint", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			ukey = []string{"ad338a8cd0332483"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.AchievesPoint(c, 3, ukey)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestDelAchieveRank(t *testing.T) {
	convey.Convey("DelAchieveRank", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelAchieveRank(c, 7)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
