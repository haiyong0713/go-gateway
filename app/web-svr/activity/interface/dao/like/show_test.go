package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestTicketAddWish(t *testing.T) {
	convey.Convey("TicketAddWish", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			ticketID = int64(2535)
			ck       = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.TicketAddWish(c, ticketID, ck)
			ctx.Convey("Then err should be nil.likes should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTicketFavCount(t *testing.T) {
	convey.Convey("TicketFavCount", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			ticketID = int64(2535)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.TicketFavCount(c, ticketID)
			ctx.Convey("Then err should be nil.likes should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%d", res)
			})
		})
	})
}
