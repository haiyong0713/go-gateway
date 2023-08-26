package dao

import (
	"context"
	"github.com/glycerine/goconvey/convey"
	"testing"
)

func TestDaoRedisLock(t *testing.T) {
	convey.Convey("redisLock", t, func(ctx convey.C) {
		err := d.RedisLock(context.Background(), "test-1", "1", 1, 1, 100)
		ctx.Convey("Then logo should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	},
	)
}

func TestDaoRedisLockErr(t *testing.T) {
	convey.Convey("redisLock", t, func(ctx convey.C) {
		err := d.RedisLock(context.Background(), "test-1", "1", 1, 1, 100)
		ctx.Convey("Then logo should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		err = d.RedisLock(context.Background(), "test-1", "1", 1, 1, 100)
		ctx.Convey("Then logo should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	},
	)
}
