package show

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestPositionCache(t *testing.T) {
	convey.Convey("Ping", t, func(ctx convey.C) {
		var (
			c   = context.TODO()
			mid = "1581871"
		)
		_, err := dao.ExistRcmmndCache(c, mid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestAddRcmmndCache(t *testing.T) {
	convey.Convey("Ping", t, func(ctx convey.C) {
		var (
			c    = context.TODO()
			mid  = "1581871"
			aids = []int64{10098955, 10098931, 10098869, 10098866, 10098865, 10098861, 10098852, 10098841, 10098839}
		)
		err := dao.AddRcmmndCache(c, mid, aids...)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestPopRcmmndCache(t *testing.T) {
	convey.Convey("Ping", t, func(ctx convey.C) {
		var (
			c   = context.TODO()
			mid = "1581871"
			cnt = 1
		)
		_, err := dao.PopRcmmndCache(c, mid, cnt)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}
