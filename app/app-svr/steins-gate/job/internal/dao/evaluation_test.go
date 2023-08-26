package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoevaluationKey(t *testing.T) {
	var (
		aid = int64(0)
	)
	convey.Convey("evaluationKey", t, func(ctx convey.C) {
		p1 := evaluationKey(aid)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoaddEvaluation(t *testing.T) {
	var (
		c          = context.Background()
		aid        = int64(0)
		evaluation = int64(0)
	)
	convey.Convey("addEvalDB", t, func(ctx convey.C) {
		err := d.addEvalDB(c, aid, evaluation)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoaddCacheEvaluation(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(4444)
		val = int64(3333)
	)
	convey.Convey("addEvalCache", t, func(ctx convey.C) {
		err := d.addEvalCache(c, aid, val)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
