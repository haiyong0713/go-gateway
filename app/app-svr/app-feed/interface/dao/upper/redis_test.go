package upper

import (
	"context"
	"testing"

	feed "git.bilibili.co/bapis/bapis-go/community/service/feed"
	"github.com/smartystreets/goconvey/convey"
)

func Test_UpItemCaches(t *testing.T) {
	convey.Convey("UpItemCaches", t, func(ctx convey.C) {
		var (
			c          = context.TODO()
			mid        int64
			start, end int
		)
		_, _, _, err := d.UpItemCaches(c, mid, start, end)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_AddUpItemCaches(t *testing.T) {
	convey.Convey("AddUpItemCaches", t, func(ctx convey.C) {
		var (
			c   = context.TODO()
			mid int64
			uis []*feed.Record
		)
		err := d.AddUpItemCaches(c, mid, uis...)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_ExpireUpItem(t *testing.T) {
	convey.Convey("ExpireUpItem", t, func(ctx convey.C) {
		var (
			c   = context.TODO()
			mid int64
		)
		_, err := d.ExpireUpItem(c, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_UnreadCountCache(t *testing.T) {
	convey.Convey("UnreadCountCache", t, func(ctx convey.C) {
		var (
			c   = context.TODO()
			mid int64
		)
		_, err := d.UnreadCountCache(c, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_AddUnreadCountCache(t *testing.T) {
	convey.Convey("AddUnreadCountCache", t, func(ctx convey.C) {
		var (
			c      = context.TODO()
			mid    int64
			unread int
		)
		err := d.AddUnreadCountCache(c, mid, unread)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
