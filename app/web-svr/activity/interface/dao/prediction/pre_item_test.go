package prediction

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawPredItems(t *testing.T) {
	convey.Convey("RawPredItems", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 2, 3}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RawPredItems(c, ids)
			ctx.Convey("Then err should be nil.ID should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

func TestItemListSet(t *testing.T) {
	convey.Convey("RawPredItems", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			id  = int64(0)
			pid = int64(3)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ItemListSet(c, id, pid)
			ctx.Convey("Then err should be nil.ID should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}
