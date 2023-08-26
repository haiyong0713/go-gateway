package handwrite

import (
	"context"
	"testing"

	mdl "go-gateway/app/web-svr/activity/job/model/handwrite"

	"github.com/glycerine/goconvey/convey"
)

func TestAddMidAward(t *testing.T) {
	convey.Convey("AddMidAward", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			midMap = map[int64]*mdl.MidAward{}
		)
		midMap[1111] = &mdl.MidAward{
			God:   1,
			Tired: 1,
			New:   1,
			Score: 222,
			Mid:   1111,
		}
		midMap[2222] = &mdl.MidAward{
			God:   1,
			Tired: 1,
			New:   1,
			Score: 222,
			Mid:   2222,
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddMidAward(c, midMap)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetMidAward(t *testing.T) {
	convey.Convey("GetMidAward", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			midMap = map[int64]*mdl.MidAward{}
		)
		midMap[1111] = &mdl.MidAward{
			God:   1,
			Tired: 1,
			New:   1,
			Score: 222,
			Mid:   1111,
		}
		midMap[2222] = &mdl.MidAward{
			God:   1,
			Tired: 1,
			New:   1,
			Score: 222,
			Mid:   2222,
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GetMidAward(c, 1111)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldResemble, midMap[1111])
			})
		})
	})
}

func TestGetMidsAward(t *testing.T) {
	convey.Convey("GetMidsAward", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			midMap = map[int64]*mdl.MidAward{}
		)
		midMap[1111] = &mdl.MidAward{
			God:   1,
			Tired: 1,
			New:   1,
			Score: 222,
			Mid:   1111,
		}
		midMap[2222] = &mdl.MidAward{
			God:   1,
			Tired: 1,
			New:   1,
			Score: 222,
			Mid:   2222,
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GetMidsAward(c, []int64{1111, 2222})
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldResemble, midMap)
			})
		})
	})
}
func TestSetAwardCount(t *testing.T) {
	convey.Convey("SetAwardCount", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		awardCount := &mdl.AwardCount{God: 111, Tired: 222, New: 2}

		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SetAwardCount(c, awardCount)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetAwardCount(t *testing.T) {
	convey.Convey("GetAwardCount", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		awardCount := &mdl.AwardCount{God: 111, Tired: 222, New: 2}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GetAwardCount(c)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldResemble, awardCount)
			})
		})
	})
}

func TestSetMidInitFans(t *testing.T) {
	convey.Convey("SetMidInitFans", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			midMap = map[int64]int64{}
		)
		midMap[1] = 333343
		midMap[5] = 4414
		midMap[8] = 44
		midMap[9] = 4411
		midMap[10] = 4411
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SetMidInitFans(c, midMap)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheActivityMember(t *testing.T) {
	convey.Convey("SetMidInitFans", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			mids = []int64{}
		)
		mids = append(mids, int64(3))
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.CacheActivityMember(c, mids)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetActivityMemberr(t *testing.T) {
	convey.Convey("SetMidInitFans", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			mids = []int64{}
		)
		mids = append(mids, int64(3))
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GetActivityMember(c)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(res, convey.ShouldResemble, mids)
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetMidInitFans(t *testing.T) {
	convey.Convey("GetMidInitFans", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetMidInitFans(c)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
