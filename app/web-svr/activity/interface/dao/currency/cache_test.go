package currency

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCurrencycurrencyKey(t *testing.T) {
	convey.Convey("currencyKey", t, func(convCtx convey.C) {
		var (
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := currencyKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencyrelationKey(t *testing.T) {
	convey.Convey("relationKey", t, func(convCtx convey.C) {
		var (
			businessID = int64(0)
			foreignID  = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := relationKey(businessID, foreignID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencyuserCurrKey(t *testing.T) {
	convey.Convey("userCurrKey", t, func(convCtx convey.C) {
		var (
			mid = int64(0)
			id  = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := userCurrKey(mid, id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}
