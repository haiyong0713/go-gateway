package currency

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCurrencyRawCurrency(t *testing.T) {
	convey.Convey("RawCurrency", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.RawCurrency(c, id)
			convCtx.Convey("Then err should be nil.data should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(data)
			})
		})
	})
}

func TestCurrencyRawRelation(t *testing.T) {
	convey.Convey("RawRelation", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			businessID = int64(1)
			foreignID  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.RawRelation(c, businessID, foreignID)
			convCtx.Convey("Then err should be nil.data should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(data)
			})
		})
	})
}

func TestCurrencyRawCurrencyUser(t *testing.T) {
	convey.Convey("RawCurrencyUser", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
			id  = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.RawCurrencyUser(c, mid, id)
			convCtx.Convey("Then err should be nil.data should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(data)
			})
		})
	})
}

func TestCurrencyCurrencySum(t *testing.T) {
	convey.Convey("CurrencySum", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			amount, err := d.CurrencySum(c, id)
			convCtx.Convey("Then err should be nil.amount should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(amount)
			})
		})
	})
}

func TestCurrencyRawCurrencyUserLog(t *testing.T) {
	convey.Convey("RawCurrencyUserLog", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
			id  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawCurrencyUserLog(c, mid, id)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestCurrencyUpUserAmount(t *testing.T) {
	convey.Convey("UpUserAmount", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			id      = int64(2)
			fromMid = int64(0)
			toMid   = int64(1)
			amount  = int64(1)
			remark  = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.UpUserAmount(c, id, fromMid, toMid, amount, remark)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
