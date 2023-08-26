package like

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/interface/model/like"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLikeactUpKey(t *testing.T) {
	Convey("actUpKey", t, func() {
		var (
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := actUpKey(mid)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestLikeactUpBySidKey(t *testing.T) {
	Convey("actUpBySidKey", t, func() {
		var (
			sid = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := actUpBySidKey(sid)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestLikeuserStateKey(t *testing.T) {
	Convey("userStateKey", t, func() {
		var (
			id  = int64(0)
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := userStateKey(id, mid)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestLikeRoundMapKey(t *testing.T) {
	Convey("RoundMapKey", t, func() {
		var (
			round = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := RoundMapKey(round)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestLikeactUpRankKey(t *testing.T) {
	Convey("actUpRankKey", t, func() {
		var (
			sid = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := actUpRankKey(sid)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestLikebuildUserDays(t *testing.T) {
	Convey("buildUserDays", t, func() {
		var (
			days  = int64(0)
			ctime = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := buildUserDays(days, ctime)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestLikeRawActUp(t *testing.T) {
	Convey("RawActUp", t, func() {
		var (
			c   = context.Background()
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.RawActUp(c, mid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeRawActUpBySid(t *testing.T) {
	Convey("RawActUpBySid", t, func() {
		var (
			c   = context.Background()
			sid = int64(1)
		)
		Convey("When everything goes positive", func() {
			res, err := d.RawActUpBySid(c, sid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeRawActUpByAid(t *testing.T) {
	Convey("RawActUpByAid", t, func() {
		var (
			c   = context.Background()
			aid = int64(880101115)
		)
		Convey("When everything goes positive", func() {
			res, err := d.RawActUpByAid(c, aid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeCacheActUp(t *testing.T) {
	Convey("CacheActUp", t, func() {
		var (
			c   = context.Background()
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheActUp(c, mid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeAddCacheActUp(t *testing.T) {
	Convey("AddCacheActUp", t, func() {
		var (
			c   = context.Background()
			mid = int64(0)
			val = &like.ActUp{}
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheActUp(c, mid, val)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeCacheActUpBySid(t *testing.T) {
	Convey("CacheActUpBySid", t, func() {
		var (
			c   = context.Background()
			sid = int64(1)
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheActUpBySid(c, sid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeCacheActUpByAid(t *testing.T) {
	Convey("CacheActUpByAid", t, func() {
		var (
			c   = context.Background()
			aid = int64(880101115)
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheActUpByAid(c, aid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}
func TestLikeAddCacheActUpBySid(t *testing.T) {
	Convey("AddCacheActUpBySid", t, func() {
		var (
			c   = context.Background()
			sid = int64(1)
			val = &like.ActUp{}
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheActUpBySid(c, sid, val)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeAddCacheActUpByAid(t *testing.T) {
	Convey("AddCacheActUpByAid", t, func() {
		var (
			c   = context.Background()
			aid = int64(0)
			val = &like.ActUp{}
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheActUpByAid(c, aid, val)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeRawUpActUserState(t *testing.T) {
	Convey("RawUpActUserState", t, func() {
		var (
			c   = context.Background()
			act = &like.ActUp{}
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.RawUpActUserState(c, act, mid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeCacheUpActUserState(t *testing.T) {
	Convey("CacheUpActUserState", t, func() {
		var (
			c   = context.Background()
			act = &like.ActUp{}
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			list, err := d.CacheUpActUserState(c, act, mid)
			Convey("Then err should be nil.list should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(list)
			})
		})
	})
}

func TestLikeAddCacheUpActUserState(t *testing.T) {
	Convey("AddCacheUpActUserState", t, func() {
		var (
			c        = context.Background()
			act      = &like.ActUp{}
			missData map[string]*like.UpActUserState
			mid      = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheUpActUserState(c, act, missData, mid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeUpActUserState(t *testing.T) {
	Convey("UpActUserState", t, func() {
		var (
			c   = context.Background()
			act = &like.ActUp{}
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.UpActUserState(c, act, mid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeSetCacheUpActUserState(t *testing.T) {
	Convey("SetCacheUpActUserState", t, func() {
		var (
			c     = context.Background()
			act   = &like.ActUp{}
			val   = &like.UpActUserState{}
			mid   = int64(0)
			round = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.SetCacheUpActUserState(c, act, val, mid, round)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeAddUserLog(t *testing.T) {
	Convey("AddUserLog", t, func() {
		var (
			c     = context.Background()
			sid   = int64(0)
			mid   = int64(0)
			bid   = int64(0)
			round = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.AddUserLog(c, sid, mid, bid, round)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeAddUserState(t *testing.T) {
	Convey("AddUserState", t, func() {
		var (
			c      = context.Background()
			sid    = int64(0)
			mid    = int64(0)
			bid    = int64(0)
			round  = int64(0)
			finish = int64(0)
			times  = int64(0)
			suffix = int64(1)
			result = ""
		)
		Convey("When everything goes positive", func() {
			err := d.AddUserState(c, sid, mid, bid, round, finish, times, suffix, result)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeUpUserState(t *testing.T) {
	Convey("UpUserState", t, func() {
		var (
			c      = context.Background()
			sid    = int64(1)
			mid    = int64(15555180)
			bid    = int64(0)
			round  = int64(0)
			finish = int64(1)
			times  = int64(1)
			suffix = int64(1)
		)
		Convey("When everything goes positive", func() {
			err := d.UpUserState(c, sid, mid, bid, round, finish, times, suffix)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeCacheUpUsersRank(t *testing.T) {
	Convey("CacheUpUsersRank", t, func() {
		var (
			c     = context.Background()
			sid   = int64(1)
			start = int(0)
			end   = int(10)
		)
		Convey("When everything goes positive", func() {
			data, err := d.CacheUpUsersRank(c, sid, start, end)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(data)
			})
		})
	})
}

func TestLikeCacheUpUserDays(t *testing.T) {
	Convey("CacheUpUserDays", t, func() {
		var (
			c   = context.Background()
			sid = int64(1)
			mid = int64(15555180)
		)
		Convey("When everything goes positive", func() {
			days, err := d.CacheUpUserDays(c, sid, mid)
			Convey("Then err should be nil.days should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(days)
			})
		})
	})
}

func TestLikeAddUpUsersRank(t *testing.T) {
	Convey("AddUpUsersRank", t, func() {
		var (
			c    = context.Background()
			sid  = int64(0)
			mid  = int64(0)
			nowT = int64(0)
			days = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.AddUpUsersRank(c, sid, mid, nowT, days)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeDeleteCacheActUpBySid(t *testing.T) {
	Convey("DeleteCacheActUpBySid", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			err := d.DeleteCacheActUpBySid(c, 0)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeDeleteCacheActUp(t *testing.T) {
	Convey("DeleteCacheActUp", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			err := d.DeleteCacheActUp(c, 0)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
