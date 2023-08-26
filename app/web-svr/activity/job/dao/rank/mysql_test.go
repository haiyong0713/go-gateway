package rank

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/job/model/handwrite"
	"go-gateway/app/web-svr/activity/job/model/rank"

	"github.com/glycerine/goconvey/convey"
)

func TestBatchAddRank(t *testing.T) {
	convey.Convey("BatchAddRank", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		dbBatch := []*rank.DB{
			{
				SID:          1111,
				Mid:          333,
				NickName:     "f",
				Rank:         1,
				Score:        222,
				State:        0,
				Batch:        20200617,
				RemarkOrigin: handwrite.Remark{Follower: 1},
			},
			{
				SID:          1111,
				Mid:          3333,
				NickName:     "f",
				Rank:         1,
				Score:        222,
				State:        0,
				Batch:        20200617,
				RemarkOrigin: handwrite.Remark{Follower: 1},
			},
			{
				SID:          1111,
				Mid:          3333,
				NickName:     "f",
				Rank:         1,
				Score:        222,
				State:        0,
				Batch:        20200617,
				RemarkOrigin: handwrite.Remark{Follower: 1},
			},
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.BatchAddRank(c, dbBatch)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetRankListByBatch(t *testing.T) {
	convey.Convey("BatchAddRank", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)

		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetRankListByBatch(c, 1111, 20200617)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetMemberRankTimes(t *testing.T) {
	convey.Convey("GetMemberRankTimes", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetMemberRankTimes(c, 1111, 20200617, 20200717, []int64{3333})
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetMemberHighest(t *testing.T) {
	convey.Convey("GetMemberHighest", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetMemberHighest(c, 1111, 20200617, 20200717, []int64{3333})
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
