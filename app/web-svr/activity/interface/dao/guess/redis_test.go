package guess

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/guess"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_AddCacheUserGuess(t *testing.T) {
	convey.Convey("AddCacheUserGuess", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			mid  = int64(1)
			miss map[int64]*guess.UserGuessLog
		)
		miss = make(map[int64]*guess.UserGuessLog)
		miss[1] = &guess.UserGuessLog{
			ID:       1,
			Mid:      1,
			DetailID: 1000,
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserGuess(c, miss, mid)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheUserGuess(t *testing.T) {
	convey.Convey("CacheUserGuess", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			mid     = int64(1)
			mainIDs = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUserGuess(c, mainIDs, mid)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_AddCacheOidMIDs(t *testing.T) {
	convey.Convey("AddCacheOidMIDs", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			oid      = int64(1)
			business = int64(1)
			miss     []*guess.MainID
		)
		miss = append(miss, &guess.MainID{ID: 1, OID: 1, IsDeleted: 1})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheOidMIDs(c, oid, miss, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheOidMIDs(t *testing.T) {
	convey.Convey("CacheOidMIDs", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			oid      = int64(1)
			business = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheOidMIDs(c, oid, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_AddCacheOidsMIDs(t *testing.T) {
	convey.Convey("AddCacheOidsMIDs", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			business = int64(1)
			miss     map[int64][]*guess.MainID
		)
		miss = make(map[int64][]*guess.MainID)
		miss[1] = append(miss[1], &guess.MainID{ID: 1, OID: 1, IsDeleted: 100})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheOidsMIDs(c, miss, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheOidsMIDs(t *testing.T) {
	convey.Convey("CacheOidsMIDs", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			buesiness = int64(1)
			oids      = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheOidsMIDs(c, oids, buesiness)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_AddCacheMDResult(t *testing.T) {
	convey.Convey("AddCacheMDResult", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			oid      = int64(1)
			business = int64(1)
			miss     *guess.MainRes
		)
		miss = &guess.MainRes{MainGuess: &guess.MainGuess{ID: 1, Business: 1, Oid: 100}}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheMDResult(c, oid, miss, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheMDResult(t *testing.T) {
	convey.Convey("CacheMDResult", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			mainID   = int64(1)
			business = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheMDResult(c, mainID, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_AddCacheMDsResult(t *testing.T) {
	convey.Convey("CacheMDsResult", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			buesiness = int64(1)
		)
		data := make(map[int64]*guess.MainRes)
		data[1] = &guess.MainRes{MainGuess: &guess.MainGuess{ID: 1, Business: 1, Oid: 100}}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheMDsResult(c, data, buesiness)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheMDsResult(t *testing.T) {
	convey.Convey("CacheMDsResult", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			buesiness = int64(1)
			mainIDs   = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheMDsResult(c, mainIDs, buesiness)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
				bs, _ := json.Marshal(res)
				fmt.Println(string(bs))
			})
		})
	})
}

func TestDao_AddCacheGuessMain(t *testing.T) {
	convey.Convey("AddCacheGuessMain", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			mainID = int64(1)
		)
		mains := &guess.MainGuess{ID: 1, Business: 1, Oid: 100}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheGuessMain(c, mainID, mains)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheGuessMain(t *testing.T) {
	convey.Convey("CacheGuessMain", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			mainID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheGuessMain(c, mainID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_AddCacheUserStat(t *testing.T) {
	convey.Convey("CacheUserGuessList", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
		)
		data := &api.UserGuessDataReply{Id: 1, Business: 1, Mid: 100}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserStat(c, mid, data, 1, 1)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheUserStat(t *testing.T) {
	convey.Convey("CacheUserGuessList", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUserStat(c, mid, 1, 1)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_DelGuessCache(t *testing.T) {
	convey.Convey("DelGuessCache", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			oid       = int64(1)
			mainID    = int64(1)
			mid       = int64(1)
			stakeType = int64(1)
			business  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelGuessCache(c, oid, business, mainID, mid, stakeType)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_AddCacheUserGuessList(t *testing.T) {
	convey.Convey("AddCacheUserGuessList", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			mid      = int64(1)
			business = int64(1)
			miss     []*guess.UserGuessLog
		)
		miss = append(miss, &guess.UserGuessLog{ID: 1, MainID: 1, Mid: 1})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserGuessList(c, mid, miss, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_CacheUserGuessList(t *testing.T) {
	convey.Convey("CacheUserGuessList", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			mid      = int64(1)
			business = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUserGuessList(c, mid, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
