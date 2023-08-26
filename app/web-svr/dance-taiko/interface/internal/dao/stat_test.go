package dao

import (
	"context"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoGatherExamples(t *testing.T) {
	var (
		c     = context.Background()
		aid   = int64(1)
		stats = []*model.Stat{
			{
				TS: 3333,
			},
			{
				TS: 4444,
			},
		}
	)
	stats[0].X = 333
	stats[1].Y = 444
	convey.Convey("GatherExamples", t, func(ctx convey.C) {
		err := d.GatherExamples(c, aid, stats)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoRedisSetUserPoints(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(1)
	)
	convey.Convey("GatherExamples", t, func(ctx convey.C) {
		err := d.RedisSetUserPoints(c, aid, map[int64]int64{11: 9999, 22: 444})
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
