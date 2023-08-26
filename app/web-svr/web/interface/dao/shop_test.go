package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestDaoShopInfo(t *testing.T) {
	convey.Convey("ShopInfo", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(27515399)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.shopURL).Reply(200).JSON(`{"code":0,"data":{"shopId":1111}}`)
			data, err := d.ShopInfo(c, mid)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}
