package currency

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/admin/model/currency"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCurrencySaveCurrency(t *testing.T) {
	Convey("SaveCurrency", t, func() {
		var (
			c   = context.Background()
			arg = &currency.SaveArg{}
		)
		Convey("When everything goes positive", func() {
			err := d.SaveCurrency(c, arg)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCurrencySaveCurrRelation(t *testing.T) {
	Convey("SaveCurrRelation", t, func() {
		var (
			c         = context.Background()
			id        = int64(0)
			isDeleted = int(0)
		)
		Convey("When everything goes positive", func() {
			err := d.SaveCurrRelation(c, id, isDeleted)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCurrencyUserCreate(t *testing.T) {
	Convey("UserCreate", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.UserCreate(c, id)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCurrencyUserLogCreate(t *testing.T) {
	Convey("UserLogCreate", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.UserLogCreate(c, id)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
