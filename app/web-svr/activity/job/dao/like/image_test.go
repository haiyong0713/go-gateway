package like

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_SetImageUpCache(t *testing.T) {
	convey.Convey("TppSetImageUpCache", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(1275392)
			day  = "20200417"
			typ  = 2
			list = []*like.ImageUp{{Mid: 1111, Score: 2222}, {Mid: 3333, Score: 4444}}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SetImageUpCache(c, sid, day, typ, list)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_ImageUpCache(t *testing.T) {
	convey.Convey("TppImageUpCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1275392)
			day = "20200417"
			typ = 2
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			list, err := d.ImageUpCache(c, sid, day, typ)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", list)
			})
		})
	})
}
