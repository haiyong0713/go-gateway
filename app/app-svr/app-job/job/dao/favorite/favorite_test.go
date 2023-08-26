package favorite

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFavoritefavStype(t *testing.T) {
	var (
		sType = ""
	)
	convey.Convey("favStype", t, func(ctx convey.C) {
		oid := favStype(sType)
		ctx.Convey("Then oid should not be nil.", func(ctx convey.C) {
			ctx.So(oid, convey.ShouldNotBeNil)
		})
	})
}

func TestFavoriteSubscriberList(t *testing.T) {
	var (
		c     = context.Background()
		sType = "weekly_selected"
		pn    = int64(0)
	)
	convey.Convey("SubscriberList", t, func(ctx convey.C) {
		reply, err := d.Subscribers(c, sType, pn)
		ctx.Convey("Then err should be nil.reply should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(reply, convey.ShouldNotBeNil)
		})
	})
}
