package archive

import (
	"context"
	"go-gateway/app/app-svr/archive/service/api"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestArchivestatRedisCache(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(22222222)
	)
	convey.Convey("statRedisCache", t, func(ctx convey.C) {
		res, err := d.StatRedisCache(c, aid)
		ctx.Convey("Then err should be nil.st should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchiveaddStatRedisCache(t *testing.T) {
	var (
		c  = context.TODO()
		st = &api.Stat{}
	)
	st.Aid = 6
	st.View = 600
	convey.Convey("addStatRedisCache", t, func(ctx convey.C) {
		err := d.addStatRedisCache(c, st)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchivestatRedisCaches(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{1, 2, 3, 4, 5, 6, 7}
	)
	convey.Convey("statRedisCaches", t, func(ctx convey.C) {
		res, miss, err := d.statRedisCaches(c, aids)
		ctx.Convey("Then err should be nil.cached,missed should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(miss, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
