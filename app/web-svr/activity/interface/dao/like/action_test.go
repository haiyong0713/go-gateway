package like

import (
	"context"
	l "go-gateway/app/web-svr/activity/interface/model/like"
	"testing"

	"fmt"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikelikeActScoreKey(t *testing.T) {
	convey.Convey("likeActScoreKey", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := likeActScoreKey(sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikelikeActScoreTypeKey(t *testing.T) {
	convey.Convey("likeActScoreTypeKey", t, func(ctx convey.C) {
		var (
			sid   = int64(10256)
			ltype = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := likeActScoreTypeKey(sid, ltype)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikelikeActKey(t *testing.T) {
	convey.Convey("likeActKey", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
			lid = int64(77)
			mid = int64(77)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := likeActKey(sid, lid, mid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikelikeLidKey(t *testing.T) {
	convey.Convey("likeLidKey", t, func(ctx convey.C) {
		var (
			oid = int64(77)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := likeLidKey(oid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikelikeCountKey(t *testing.T) {
	convey.Convey("likeCountKey", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := likeCountKey(sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeActInfos(t *testing.T) {
	convey.Convey("LikeActInfos", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			lids = []int64{1, 2}
			mid  = int64(77)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			likeActs, err := d.LikeActInfos(c, lids, mid)
			ctx.Convey("Then err should be nil.likeActs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(likeActs, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeStoryLikeActSum(t *testing.T) {
	convey.Convey("StoryLikeActSum", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10256)
			mid   = int64(77)
			stime = "2018-10-16 00:00:00"
			etime = "2018-10-16 23:59:59"
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.StoryLikeActSum(c, sid, mid, stime, etime)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeStoryEachLikeAct(t *testing.T) {
	convey.Convey("StoryEachLikeAct", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10296)
			mid   = int64(216761)
			lid   = int64(13538)
			stime = "2018-10-17 00:00:00"
			etime = "2018-10-17 23:59:59"
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.StoryEachLikeAct(c, sid, mid, lid, stime, etime)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%d", res)
			})
		})
	})
}

func TestLikeSetRedisCache(t *testing.T) {
	convey.Convey("SetRedisCache", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			sid      = int64(10296)
			lid      = int64(13538)
			score    = int64(10)
			likeType = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetRedisCache(c, sid, lid, score, likeType)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeRedisCache(t *testing.T) {
	convey.Convey("RedisCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10296)
			start = int64(0)
			end   = int64(100)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RedisCache(c, sid, 0, start, end)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestLikeLikeActZscore(t *testing.T) {
	convey.Convey("LikeActZscore", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10296)
			lid = int64(13528)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeActZscore(c, sid, lid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%d", res)
			})
		})
	})
}

func TestLikeSetInitializeLikeCache(t *testing.T) {
	convey.Convey("SetInitializeLikeCache", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			sid        = int64(10256)
			lidLikeAct = map[int64]int64{77: 1, 88: 2}
			typeLike   = map[int64]int64{77: 1, 88: 2}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetInitializeLikeCache(c, sid, lidLikeAct, typeLike)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeLikeActAdd(t *testing.T) {
	convey.Convey("LikeActAdd", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			likeAct = &l.Action{Sid: 10256, Lid: 77, Action: 1, IPv6: make([]byte, 0)}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			id, err := d.LikeActAdd(c, likeAct)
			ctx.Convey("Then err should be nil.id should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(id, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeActLidCounts(t *testing.T) {
	convey.Convey("LikeActLidCounts", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			lids = []int64{2354, 2355}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeActLidCounts(c, lids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestLikeLikeActs(t *testing.T) {
	convey.Convey("LikeActs", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(10256)
			mid  = int64(77)
			lids = []int64{1188}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeActs(c, sid, mid, lids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeCacheLikeActs(t *testing.T) {
	convey.Convey("CacheLikeActs", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(10256)
			mid  = int64(77)
			lids = []int64{1185, 1186, 1188, 1188, 1187}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheLikeActs(c, sid, mid, lids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestLikeAddCacheLikeActs(t *testing.T) {
	convey.Convey("AddCacheLikeActs", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			sid    = int64(10256)
			mid    = int64(77)
			values = map[int64]int{1185: 1, 1186: 2, 1187: 1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCacheLikeActs(c, sid, mid, values)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestSetLikesReload(t *testing.T) {
	convey.Convey("SetLikesReload", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10256)
			lid = int64(1185)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetLikesReload(c, lid, sid, 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDelLikeListLikes(t *testing.T) {
	convey.Convey("DelLikeListLikes", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(10256)
			item = make([]*l.Item, 0, 2)
		)
		item = append(item, &l.Item{ID: 1185, Sid: 10256})
		item = append(item, &l.Item{ID: 1256, Sid: 10256})
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.DelLikeListLikes(c, sid, item)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestSetLikeLimitNum(t *testing.T) {
	convey.Convey("SetLikeLimitNum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10256)
			mid = int64(48875654)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetLikeLimitNum(c, sid, mid, 1)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetLikeLimitNum(t *testing.T) {
	convey.Convey("GetLikeLimitNum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10256)
			mid = int64(48875654)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.GetLikeLimitNum(c, sid, mid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestIncrLikeExtraTimes(t *testing.T) {
	convey.Convey("IncrLikeExtraTimes", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.IncrLikeExtraTimes(c, 10555, 88895359, 1)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawLikeActLids(t *testing.T) {
	convey.Convey("RawLikeActLids", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10552)
			mid = int64(27515240)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.RawLikeActLids(c, sid, mid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestIncrLikeExtendToken(t *testing.T) {
	convey.Convey("IncrLikeExtendToken", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.IncrLikeExtendToken(c, 2, 12221, 3, "11test441")
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeActLids(t *testing.T) {
	convey.Convey("LikeActLids", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10552)
			mid = int64(27515240)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.LikeActLids(c, sid, mid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestLikeTextOnly(t *testing.T) {
	convey.Convey("TextOnly", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10552)
			mid = int64(27515240)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.TextOnly(c, sid, mid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestLikeActUserAward(t *testing.T) {
	convey.Convey("ActUserAward", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.ActUserAward(c, id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestLikeLikes(t *testing.T) {
	convey.Convey("Likes", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.Likes(c, ids)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestLikeActSubjects(t *testing.T) {
	convey.Convey("ActSubjects", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{10552}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.ActSubjects(c, ids)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}
