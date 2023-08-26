package like

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeipRequestKey(t *testing.T) {
	convey.Convey("ipRequestKey", t, func(ctx convey.C) {
		var (
			ip = "10.256.8.3"
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := ipRequestKey(ip)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikelikeListCtimeKey(t *testing.T) {
	convey.Convey("likeListCtimeKey", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := likeListCtimeKey(sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeListRandomKey(t *testing.T) {
	convey.Convey("likeListCtimeKey", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			p1 := likeListRandomKey(sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeStochasticKey(t *testing.T) {
	convey.Convey("likeListCtimeKey", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			p1 := likeStochasticKey(sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikelikeListTypeCtimeKey(t *testing.T) {
	convey.Convey("likeListTypeCtimeKey", t, func(ctx convey.C) {
		var (
			types = int64(1)
			sid   = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := likeListTypeCtimeKey(types, sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeyLikeTag(t *testing.T) {
	convey.Convey("keyLikeTag", t, func(ctx convey.C) {
		var (
			sid   = int64(10256)
			tagID = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyLikeTag(sid, tagID)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeyLikeTagCounts(t *testing.T) {
	convey.Convey("keyLikeTagCounts", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyLikeTagCounts(sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeyLikeRegion(t *testing.T) {
	convey.Convey("keyLikeRegion", t, func(ctx convey.C) {
		var (
			sid      = int64(10256)
			regionID = int32(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyLikeRegion(sid, regionID)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeyStoryLikeKey(t *testing.T) {
	convey.Convey("keyStoryLikeKey", t, func(ctx convey.C) {
		var (
			sid   = int64(10256)
			mid   = int64(1)
			daily = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyStoryLikeKey(sid, mid, daily)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeyStoryEachLike(t *testing.T) {
	convey.Convey("keyStoryEachLike", t, func(ctx convey.C) {
		var (
			sid   = int64(10256)
			mid   = int64(1)
			lid   = int64(1)
			daily = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyStoryEachLike(sid, mid, lid, daily)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeykeyStoryExtendToken(t *testing.T) {
	convey.Convey("keyStoryExtendToken", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
			mid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyStoryExtendToken(sid, mid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeykeyStoryExtraLike(t *testing.T) {
	convey.Convey("keyStoryExtraLike", t, func(ctx convey.C) {
		var (
			sid = int64(10256)
			mid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyStoryExtraLike(sid, mid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeykeyStoryExtendInfo(t *testing.T) {
	convey.Convey("keyStoryEachLike", t, func(ctx convey.C) {
		var (
			sid   = int64(10256)
			token = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyStoryExtendInfo(sid, token)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeList(t *testing.T) {
	convey.Convey("LikeList", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			ns, err := d.LikeList(c, sid)
			ctx.Convey("Then err should be nil.ns should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(ns, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeTagCache(t *testing.T) {
	convey.Convey("LikeTagCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10256)
			tagID = int64(1)
			start = int(1)
			end   = int(2)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			likes, err := d.LikeTagCache(c, sid, tagID, start, end)
			ctx.Convey("Then err should be nil.likes should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", likes)
			})
		})
	})
}

func TestLikeLikeTagCnt(t *testing.T) {
	convey.Convey("LikeTagCnt", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10256)
			tagID = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			count, err := d.LikeTagCnt(c, sid, tagID)
			ctx.Convey("Then err should be nil.count should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(count, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeRegionCache(t *testing.T) {
	convey.Convey("LikeRegionCache", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			sid      = int64(10256)
			regionID = int32(1)
			start    = int(1)
			end      = int(2)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			likes, err := d.LikeRegionCache(c, sid, regionID, start, end)
			ctx.Convey("Then err should be nil.likes should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", likes)
			})
		})
	})
}

func TestLikeLikeRegionCnt(t *testing.T) {
	convey.Convey("LikeRegionCnt", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			sid      = int64(10256)
			regionID = int32(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			count, err := d.LikeRegionCnt(c, sid, regionID)
			ctx.Convey("Then err should be nil.count should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(count, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeSetLikeRegionCache(t *testing.T) {
	convey.Convey("SetLikeRegionCache", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			sid      = int64(10256)
			regionID = int32(1)
			likes    = []*like.Item{{Sid: 10256, Wid: 1, Mid: 44}}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetLikeRegionCache(c, sid, regionID, likes)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeSetTagLikeCountsCache(t *testing.T) {
	convey.Convey("SetTagLikeCountsCache", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			sid    = int64(10256)
			counts = map[int64]int32{1: 2}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetTagLikeCountsCache(c, sid, counts)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeTagLikeCountsCache(t *testing.T) {
	convey.Convey("TagLikeCountsCache", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			sid    = int64(10256)
			tagIDs = []int64{1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			counts, err := d.TagLikeCountsCache(c, sid, tagIDs)
			ctx.Convey("Then err should be nil.counts should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(counts, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeRawLike(t *testing.T) {
	convey.Convey("RawLike", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RawLike(c, id)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeListMoreLid(t *testing.T) {
	convey.Convey("LikeListMoreLid", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			lid = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeListMoreLid(c, lid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikesBySid(t *testing.T) {
	convey.Convey("LikesBySid", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			lid = int64(77)
			sid = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikesBySid(c, lid, sid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeListCtime(t *testing.T) {
	convey.Convey("LikeListCtime", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10256)
			items = []*like.Item{{Sid: 10256, Wid: 55, Mid: 234}}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.LikeListCtime(c, sid, items)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeDelLikeListCtime(t *testing.T) {
	convey.Convey("DelLikeListCtime", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10256)
			items = []*like.Item{{Sid: 10256, Wid: 55, Mid: 234}}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.DelLikeListCtime(c, sid, items)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeLikeMaxID(t *testing.T) {
	convey.Convey("LikeMaxID", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeMaxID(c)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestLikeStoryLikeSum(t *testing.T) {
	convey.Convey("StoryLikeSum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10256)
			mid = int64(77)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.StoryLikeSum(c, sid, mid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeIncrStoryLikeSum(t *testing.T) {
	convey.Convey("IncrStoryLikeSum", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10256)
			mid   = int64(77)
			score = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.IncrStoryLikeSum(c, sid, mid, score)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeSetLikeSum(t *testing.T) {
	convey.Convey("SetLikeSum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10256)
			mid = int64(77)
			sum = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetLikeSum(c, sid, mid, sum)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				fmt.Printf("%v", err)
			})
		})
	})
}

func TestLikeStoryEachLikeSum(t *testing.T) {
	convey.Convey("StoryEachLikeSum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10296)
			mid = int64(216761)
			lid = int64(13538)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.StoryEachLikeSum(c, sid, mid, lid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%d", res)
			})
		})
	})
}

func TestBatchStoryEachLikeSum(t *testing.T) {
	convey.Convey("BatchStoryEachLikeSum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10296)
			mid = int64(216761)
			lid = []int64{13538}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.BatchStoryEachLikeSum(c, sid, mid, lid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%d", res)
			})
		})
	})
}

func TestLikeIncrStoryEachLikeAct(t *testing.T) {
	convey.Convey("IncrStoryEachLikeAct", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10256)
			mid   = int64(77)
			lid   = int64(77)
			score = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.IncrStoryEachLikeAct(c, sid, mid, lid, score)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeSetEachLikeSum(t *testing.T) {
	convey.Convey("SetEachLikeSum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10256)
			mid = int64(77)
			lid = int64(77)
			sum = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetEachLikeSum(c, sid, mid, lid, sum)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				fmt.Printf("%v", err)
			})
		})
	})
}

func TestLikeCtime(t *testing.T) {
	convey.Convey("SetEachLikeSum", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10365)
			start = int64(1)
			end   = int64(100)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeCtime(c, sid, 0, start, end)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

func TestLikeRandom(t *testing.T) {
	convey.Convey("SetEachLikeSum", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10365)
			start = int64(1)
			end   = int64(100)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheActRandom(c, sid, 0, start, end)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

func TestLikeRandomCount(t *testing.T) {
	convey.Convey("SetEachLikeSum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10365)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.LikeRandomCount(c, sid, 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

func TestSetLikeRandom(t *testing.T) {
	convey.Convey("SetEachLikeSum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10365)
			ids = []int64{2354, 2355}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.AddCacheActRandom(c, sid, ids, 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeCount(t *testing.T) {
	convey.Convey("SetEachLikeSum", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10365)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeCount(c, sid, 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

func TestDao_SourceItemData(t *testing.T) {
	convey.Convey("test group item data", t, func(ctx convey.C) {
		sid := int64(37)
		data, err := d.SourceItemData(context.Background(), sid)
		convey.So(err, convey.ShouldBeNil)
		str, _ := json.Marshal(data)
		convey.Printf("%+v", string(str))
	})
}

func TestDao_ListFromES(t *testing.T) {
	convey.Convey("test group item data", t, func(ctx convey.C) {
		sid := int64(10533)
		ps := 15
		pn := 1
		data, err := d.ListFromES(context.Background(), sid, "", ps, pn, time.Now().Unix(), 25)
		convey.So(err, convey.ShouldBeNil)
		if data != nil {
			for _, v := range data.List {
				convey.Printf(" %+v ", v.Item)
			}
		}
	})
}

func TestDao_MyListFromEs(t *testing.T) {
	convey.Convey("MyListFromEs", t, func(ctx convey.C) {
		sid := int64(10572)
		mid := int64(27515234)
		ps := 50
		pn := 1
		isOriginal := 0
		data, err := d.MyListFromEs(context.Background(), sid, mid, "id", ps, pn, isOriginal)
		convey.So(err, convey.ShouldBeNil)
		convey.So(data, convey.ShouldNotBeNil)
		for _, v := range data.List {
			convey.Printf(" %+v ", v.Item)
		}
	})
}

func TestDao_AllListFromEs(t *testing.T) {
	convey.Convey("AllListFromEs", t, func(ctx convey.C) {
		sids := []int64{10567, 10572, 10573, 10570}
		mid := int64(27515234)
		ps := 5
		pn := 1
		isOriginal := 0
		data, err := d.AllListFromEs(context.Background(), sids, mid, "ctime", ps, pn, isOriginal)
		convey.So(err, convey.ShouldBeNil)
		convey.So(data, convey.ShouldNotBeNil)
		for _, v := range data.List {
			convey.Printf(" %+v ", v.Item)
		}
	})
}

func TestDao_MyListTotalStateFromEs(t *testing.T) {
	convey.Convey("MyListTotalStateFromEs", t, func(ctx convey.C) {
		sid := int64(10517)
		mid := int64(37924693)
		isOriginal := 1
		data, err := d.MyListTotalStateFromEs(context.Background(), sid, mid, isOriginal)
		convey.So(err, convey.ShouldBeNil)
		convey.So(data, convey.ShouldNotBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_StateModify(t *testing.T) {
	convey.Convey("StateModify", t, func(ctx convey.C) {
		lid := int64(1185)
		mid := int64(15555180)
		_, err := d.StateModify(context.Background(), lid, mid, 1)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_MultiTags(t *testing.T) {
	convey.Convey("test group item data", t, func(ctx convey.C) {
		wids := []int64{10109984}
		data, err := d.MultiTags(context.Background(), wids)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_OidInfoFromES(t *testing.T) {
	convey.Convey("test group item data", t, func(ctx convey.C) {
		oids := []int64{11, 21}
		stype := 1
		ps := 50
		pn := 1
		data, err := d.OidInfoFromES(context.Background(), oids, stype, ps, pn)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_LikeStochastic(t *testing.T) {
	convey.Convey("LikeStochastic", t, func(ctx convey.C) {
		sid := int64(1)
		data, err := d.CacheActStochastic(context.Background(), sid, 0)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_SetLikeStochastic(t *testing.T) {
	convey.Convey("SetLikeStochastic", t, func(ctx convey.C) {
		sid := int64(1)
		ids := []int64{1, 2, 3, 4}
		err := d.AddCacheActStochastic(context.Background(), sid, ids, 0)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_TxAddLike(t *testing.T) {
	convey.Convey("TxAddLike", t, func(ctx convey.C) {
		c := context.Background()
		tx, _ := d.db.Begin(c)
		item := &like.Item{}
		data, err := d.TxAddLike(c, tx, item)
		tx.Rollback()
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_RawLikeExtendToken(t *testing.T) {
	convey.Convey("RawLikeExtendToken", t, func(ctx convey.C) {
		data, err := d.RawLikeExtendToken(context.Background(), 1, 1)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_RawLikeExtendInfo(t *testing.T) {
	convey.Convey("RawLikeExtendInfo", t, func(ctx convey.C) {
		data, err := d.RawLikeExtendInfo(context.Background(), 1, "")
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_RawLikeExtendTimes(t *testing.T) {
	convey.Convey("RawLikeExtraTimes", t, func(ctx convey.C) {
		data, err := d.RawLikeExtraTimes(context.Background(), 10555, 88895359)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_CacheLikeExtraTimes(t *testing.T) {
	convey.Convey("CacheLikeExtraTimes", t, func(ctx convey.C) {
		data, err := d.CacheLikeExtraTimes(context.Background(), 1, 1, 0, -1)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_AddCacheLikeExtraTimes(t *testing.T) {
	convey.Convey("AddCacheLikeExtraTimes", t, func(ctx convey.C) {
		val := []*like.ExtraTimesDetail{{ID: 1}}
		err := d.AddCacheLikeExtraTimes(context.Background(), 1, 1, val)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_RawActStochastic(t *testing.T) {
	convey.Convey("TxAddLike", t, func(ctx convey.C) {
		c := context.Background()
		_, err := d.RawActStochastic(c, 10533, 0)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_RawActRandom(t *testing.T) {
	convey.Convey("RawActRandom", t, func(ctx convey.C) {
		c := context.Background()
		_, _, err := d.RawActRandom(c, 10533, 0, 0, 1)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_RawLikeTotal(t *testing.T) {
	convey.Convey("RawLikeTotal", t, func(ctx convey.C) {
		c := context.Background()
		_, err := d.RawLikeTotal(c, 10533)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_AddCacheLikeTotal(t *testing.T) {
	convey.Convey("AddCacheLikeTotal", t, func(ctx convey.C) {
		c := context.Background()
		err := d.AddCacheLikeTotal(c, 10533, 48)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_CacheLikeTotal(t *testing.T) {
	convey.Convey("RawLikeTotal", t, func(ctx convey.C) {
		c := context.Background()
		_, err := d.CacheLikeTotal(c, 10533)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_AddLikeExtraTimes(t *testing.T) {
	convey.Convey("AddLikeExtraTimes", t, func(ctx convey.C) {
		val := &like.ExtraTimesDetail{ID: 1}
		err := d.AddLikeExtraTimes(context.Background(), 1, 1, val)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_IncrCacheLikeTotal(t *testing.T) {
	convey.Convey("IncrCacheLikeTotal", t, func(ctx convey.C) {
		c := context.Background()
		err := d.IncrCacheLikeTotal(c, 10533)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_CacheLikeExtendInfo(t *testing.T) {
	convey.Convey("CacheLikeExtendInfo", t, func(ctx convey.C) {
		data, err := d.CacheLikeExtendInfo(context.Background(), 1, "")
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_AddCacheLikeExtendInfo(t *testing.T) {
	convey.Convey("AddCacheLikeExtendInfo", t, func(ctx convey.C) {
		val := &like.ExtendTokenDetail{ID: 1}
		err := d.AddCacheLikeExtendInfo(context.Background(), 1, val, "")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_CacheLikeExtendToken(t *testing.T) {
	convey.Convey("CacheLikeExtendToken", t, func(ctx convey.C) {
		data, err := d.CacheLikeExtendToken(context.Background(), 1, 1)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_AddCacheLikeExtendToken(t *testing.T) {
	convey.Convey("AddCacheLikeExtendToken", t, func(ctx convey.C) {
		val := &like.ExtendTokenDetail{ID: 1}
		err := d.AddCacheLikeExtendToken(context.Background(), 1, val, 1)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_LikeExtendInfo(t *testing.T) {
	convey.Convey("LikeExtendInfo", t, func(ctx convey.C) {
		data, err := d.LikeExtendInfo(context.Background(), 1, "")
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_LikeExtendToken(t *testing.T) {
	convey.Convey("LikeExtendToken", t, func(ctx convey.C) {
		data, err := d.LikeExtendToken(context.Background(), 1, 1)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v", data)
	})
}

func TestDao_RawLikeCheck(t *testing.T) {
	convey.Convey("RawLikeCheck", t, func(ctx convey.C) {
		c := context.Background()
		_, err := d.RawLikeCheck(c, 10200322, 10533)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_RawTextOnly(t *testing.T) {
	convey.Convey("RawTextOnly", t, func(ctx convey.C) {
		c := context.Background()
		_, err := d.RawTextOnly(c, 10533, 10200322)
		convey.So(err, convey.ShouldBeNil)
	})
}
