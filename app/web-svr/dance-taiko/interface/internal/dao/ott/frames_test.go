package ott

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/glycerine/goconvey/convey"
)

func TestDao_AddCacheKeyFrames(t *testing.T) {
	var (
		c      = context.Background()
		cid    = int64(10099306)
		frames = &model.FramesCache{}
	)
	convey.Convey("TestDao_AddCacheKeyFrames", t, func(ctx convey.C) {
		err := d.addCacheKeyFrames(c, cid, frames)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_CacheKeyFrames(t *testing.T) {
	var (
		c   = context.Background()
		cid = int64(10099306)
	)
	convey.Convey("TestDao_CacheKeyFrames", t, func(ctx convey.C) {
		res, err := d.cacheKeyFrames(c, cid)
		ctx.So(res, convey.ShouldNotBeNil)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestDao_RawKeyFrames(t *testing.T) {
	var (
		c   = context.Background()
		cid = int64(10099306)
	)
	convey.Convey("TestDao_RawKeyFrames", t, func(ctx convey.C) {
		res, err := d.rawKeyFrames(c, cid)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
	})
}
