package like

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeRawLikeContent(t *testing.T) {
	convey.Convey("RawLikeContent", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 2}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RawLikeContent(c, ids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_TxAddContent(t *testing.T) {
	convey.Convey("TxAddContent", t, func(ctx convey.C) {
		var err error
		c := context.Background()
		tx, _ := d.db.Begin(c)
		item := &like.LikeContent{ID: 1454, Message: "messga", IPv6: []byte{}}
		d.TxAddContent(c, tx, item)
		tx.Rollback()
		convey.So(err, convey.ShouldBeNil)
	})
}
