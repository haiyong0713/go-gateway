package bplus

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/model"

	. "github.com/smartystreets/goconvey/convey"
)

// TestKeyContributeAttr dao ut.
func TestKeyContributeAttr(t *testing.T) {
	Convey("keyContributeAttr", t, func(ctx C) {
		key := keyContributeAttr(int64(123456))
		ctx.Convey("key should not be equal to 123456", func(ctx C) {
			ctx.So(key, ShouldEqual, "cba_123456")
		})
	})
}

// TestKeyContribute dao ut.
func TestKeyContribute(t *testing.T) {
	Convey("keyContribute", t, func(ctx C) {
		key := keyContribute(int64(123456))
		ctx.Convey("key should not be equal to 123456", func(ctx C) {
			ctx.So(key, ShouldEqual, "cb_123456")
		})
	})
}

// TestKeyContributeAttrCooperation dao ut.
func TestKeyContributeAttrCooperation(t *testing.T) {
	Convey("keyContributeAttr", t, func(ctx C) {
		key := keyContributeAttrCooperation(int64(123456))
		ctx.Convey("key should not be equal to 123456", func(ctx C) {
			ctx.So(key, ShouldEqual, "cbacoop_123456")
		})
	})
}

// TestKeyContributeCooperation dao ut.
func TestKeyContributeCooperation(t *testing.T) {
	Convey("keyContribute", t, func(ctx C) {
		key := keyContributeCooperation(int64(123456))
		ctx.Convey("key should not be equal to 123456", func(ctx C) {
			ctx.So(key, ShouldEqual, "cbcoop_123456")
		})
	})
}

// TestRangeContributeCache dao ut.
func TestRangeContributeCache(t *testing.T) {
	Convey("RangeContributeCache", t, func(ctx C) {
		_, err := dao.RangeContributeCache(context.Background(), 123456, 1, 20, false)
		ctx.Convey("Then err should not be nil.", func(ctx C) {
			ctx.So(err, ShouldNotBeNil)
		})
	})
}

// TestRangeContributionCache dao ut.
func TestRangeContributionCache(t *testing.T) {
	Convey("RangeContributionCache", t, func(ctx C) {
		var cursor = &model.Cursor{}
		_, err := dao.RangeContributionCache(context.Background(), 123456, cursor)
		ctx.Convey("Then err should not be nil.", func(ctx C) {
			ctx.So(err, ShouldNotBeNil)
		})
	})
}

// TestAttrCache dao ut.
func TestAttrCache(t *testing.T) {
	Convey("AttrCache", t, func(ctx C) {
		_, err := dao.AttrCache(context.Background(), 123456, false)
		ctx.Convey("Then err should not be nil.", func(ctx C) {
			ctx.So(err, ShouldNotBeNil)
		})
	})
}
