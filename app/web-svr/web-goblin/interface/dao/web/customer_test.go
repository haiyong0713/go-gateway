package web

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/web-goblin/interface/model/web"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoCusCenter(t *testing.T) {
	convey.Convey("CusCenter", t, func(c convey.C) {
		var (
			ctx = context.Background()
		)
		res, err := d.CusCenter(ctx)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
	})
}

func TestDaoCustomerCache(t *testing.T) {
	convey.Convey("CustomerCache", t, func(c convey.C) {
		var (
			ctx = context.Background()
		)
		res, err := d.CustomerCache(ctx)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
	})
}

func TestDao_SetCustomerCache(t *testing.T) {
	convey.Convey("SetCustomerCache", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			rs map[string]*web.CustomerCenter
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			rs = make(map[string]*web.CustomerCenter, 1)
			rs["_hint"] = &web.CustomerCenter{}
			err := d.SetCustomerCache(c, rs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
