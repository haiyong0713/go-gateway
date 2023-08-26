package unicom

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestUnicomOrdersUserFlow(t *testing.T) {
	var (
		ctx     = context.Background()
		usermob = ""
	)
	convey.Convey("OrdersUserFlow", t, func(c convey.C) {
		res, err := d.OrdersUserFlow(ctx, usermob)
		c.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomBindAll(t *testing.T) {
	var (
		ctx   = context.Background()
		start = int(0)
		end   = int(1)
	)
	convey.Convey("BindAll", t, func(c convey.C) {
		res, err := d.BindAll(ctx, start, end)
		c.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomUserBind(t *testing.T) {
	var (
		ctx = context.Background()
		mid = int64(27515399)
	)
	convey.Convey("UserBind", t, func(c convey.C) {
		res, err := d.UserBind(ctx, mid)
		c.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomIPSync(t *testing.T) {
	var (
		ctx = context.Background()
	)
	convey.Convey("IPSync", t, func(c convey.C) {
		res, err := d.IPSync(ctx)
		c.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomBeginTran(t *testing.T) {
	var (
		ctx = context.Background()
	)
	convey.Convey("BeginTran", t, func(c convey.C) {
		tx, err := d.BeginTran(ctx)
		c.Convey("Then err should be nil.tx should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(tx, convey.ShouldNotBeNil)
		})
	})
}
