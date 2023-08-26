package pay

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_PayTransferInner(t *testing.T) {
	convey.Convey("PayTransferInner", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			pt = &PayTransferInner{
				TraceID:      "1591265303946",
				UID:          1111119436,
				OrderNo:      "e4d36a4b8ce951ebe978ea8849eb3fb1", // 业务方转入红包的订单id（通过该字段保持幂等）
				TransBalance: 30,                                 // 分转成元
				TransDesc:    "小程序抽奖",                            // 红包名称
				StartTme:     1591442245226,                      // 红包解冻时间
				Timestamp:    1591355845226,                      // 当前时间毫秒值
			}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.PayTransferInner(c, pt)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(data, convey.ShouldNotBeNil)
				convCtx.Printf("%+v", data)
			})
		})
	})
}
