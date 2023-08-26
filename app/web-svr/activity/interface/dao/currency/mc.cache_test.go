package currency

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/model/currency"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCurrencyCacheCurrency(t *testing.T) {
	convey.Convey("CacheCurrency", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheCurrency(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestCurrencyAddCacheCurrency(t *testing.T) {
	convey.Convey("AddCacheCurrency", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(2)
			val = &currency.Currency{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheCurrency(c, id, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCurrencyCacheRelation(t *testing.T) {
	convey.Convey("CacheRelation", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			id        = int64(0)
			foreignID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheRelation(c, id, foreignID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestCurrencyAddCacheRelation(t *testing.T) {
	convey.Convey("AddCacheRelation", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			id        = int64(0)
			val       = &currency.CurrencyRelation{}
			foreignID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheRelation(c, id, val, foreignID)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCurrencyCacheCurrencyUser(t *testing.T) {
	convey.Convey("CacheCurrencyUser", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			id     = int64(1)
			currID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheCurrencyUser(c, id, currID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestCurrencyAddCacheCurrencyUser(t *testing.T) {
	convey.Convey("AddCacheCurrencyUser", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			id     = int64(2)
			val    = &currency.CurrencyUser{}
			currID = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheCurrencyUser(c, id, val, currID)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCurrencyDelCacheCurrencyUser(t *testing.T) {
	convey.Convey("DelCacheCurrencyUser", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			id     = int64(2)
			currID = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheCurrencyUser(c, id, currID)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
