package lottery

import (
	"context"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLotteryGetAddressByID(t *testing.T) {
	Convey("GetAddressByID", t, func() {
		var (
			c   = context.Background()
			id  = int64(1002)
			uid = int(88889062)
		)
		Convey("When everything goes positive", func() {
			_, err := d.GetAddressByID(c, id, uid)
			Convey("Then err should be nil.addr should not be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLotteryGetVIPInfo(t *testing.T) {
	Convey("GetVIPInfo", t, func() {
		var (
			c      = context.Background()
			id     = ""
			cookie = ""
		)
		Convey("When everything goes positive", func() {
			_, err := d.GetVIPInfo(c, id, cookie)
			Convey("Then err should be nil.info should not be nil.", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestLotteryGetCouponInfo(t *testing.T) {
	Convey("GetCouponInfo", t, func() {
		var (
			c      = context.Background()
			token  = ""
			cookie = ""
		)
		Convey("When everything goes positive", func() {
			_, err := d.GetCouponInfo(c, token, cookie)
			Convey("Then err should be nil.couponInfo should not be nil.", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestLotterysign(t *testing.T) {
	Convey("sign", t, func() {
		var (
			params = &url.Values{}
			key    = "token"
			value  = "233"
		)
		Convey("When everything goes positive", func() {
			query := d.sign(params, key, value)
			Convey("Then query should not be nil.", func() {
				So(query, ShouldNotBeNil)
			})
		})
	})
}
