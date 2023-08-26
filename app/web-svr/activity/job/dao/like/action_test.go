package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeBatchLikeActSum(t *testing.T) {
	convey.Convey("BatchLikeActSum", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			lids = []int64{13511, 13512, 13510}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			_, err := d.BatchLikeActSum(c, lids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeActList(t *testing.T) {
	convey.Convey("LikeActList", t, func(ctx convey.C) {
		var (
			c           = context.Background()
			lid   int64 = 13511
			minID int64 = 100
			limit int64 = 10
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.LikeActList(c, lid, minID, limit)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", res)
			})
		})
	})
}

func TestLikeActState(t *testing.T) {
	convey.Convey("LikeActState", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			sid  int64 = 10457
			mid  int64 = 15555180
			lids       = []int64{3044}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.LikeActState(c, sid, mid, lids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", res)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
