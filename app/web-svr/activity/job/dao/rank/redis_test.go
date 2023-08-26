package rank

import (
	"context"
	"testing"

	rankMdl "go-gateway/app/web-svr/activity/job/model/rank"

	"github.com/glycerine/goconvey/convey"
)

func TestSetRank(t *testing.T) {
	convey.Convey("SetRank", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			rank = []*rankMdl.Redis{}
		)
		rank = append(rank, &rankMdl.Redis{
			Mid: 1, Rank: 2, Score: 1,
		})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SetRank(c, "handwrite", rank)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetRank(t *testing.T) {
	convey.Convey("GetRank", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			rank = []*rankMdl.Redis{}
		)
		rank = append(rank, &rankMdl.Redis{
			Mid: 1, Rank: 2, Score: 1,
		})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GetRank(c, "handwrite")
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldResemble, rank)
			})
		})
	})
}
