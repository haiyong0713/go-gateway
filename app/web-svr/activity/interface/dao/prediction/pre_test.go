package prediction

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawPredictions(t *testing.T) {
	convey.Convey("RawPredictions", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 3}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RawPredictions(c, ids)
			ctx.Convey("Then err should be nil.ID should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

func TestListSet(t *testing.T) {
	convey.Convey("ListSet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			id  = int64(0)
			sid = int64(10292)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ListSet(c, id, sid)
			ctx.Convey("Then err should be nil.ID should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}
