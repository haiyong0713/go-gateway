package newstar

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/interface/model/newstar"

	"github.com/glycerine/goconvey/convey"
)

func TestDao_CacheCreationByMid(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("CacheCreationByMid", t, func(ctx convey.C) {
		_, err := d.CacheCreationByMid(c, 88895033, "newstar_first")
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_AddCacheCreationByMid(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("AddCacheCreationByMid", t, func(ctx convey.C) {
		data := newstar.Newstar{
			ID:          111,
			ActivityUID: "newstar_first",
			Mid:         88895033,
		}
		err := d.AddCacheCreationByMid(c, 88895033, "newstar_first", &data)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_CacheInvites(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("CacheInvites", t, func(ctx convey.C) {
		res, err := d.CacheInvites(c, 88895033, "newstar_first")
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_AddCacheInvites(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("AddCacheInvites", t, func(ctx convey.C) {
		data := newstar.Newstar{
			ID:          111,
			ActivityUID: "newstar_first",
			Mid:         88895033,
		}
		err := d.AddCacheInvites(c, 88895033, "newstar_first", []*newstar.Newstar{&data})
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
