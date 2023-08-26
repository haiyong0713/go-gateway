package currency

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCurrencyCurrency(t *testing.T) {
	convey.Convey("Currency", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Currency(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencyRelation(t *testing.T) {
	convey.Convey("Relation", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			businessID = int64(1)
			foreignID  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Relation(c, businessID, foreignID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCurrencyCurrencyUser(t *testing.T) {
	convey.Convey("CurrencyUser", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
			id  = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CurrencyUser(c, mid, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
