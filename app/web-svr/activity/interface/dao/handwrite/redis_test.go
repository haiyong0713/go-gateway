package handwrite

import (
	"context"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestGetMidAward(t *testing.T) {
	convey.Convey("GetMidAward", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)

		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetMidAward(c, 1111)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetAwardCount(t *testing.T) {
	convey.Convey("GetAwardCount", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)

		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetAwardCount(c)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddTimeLock(t *testing.T) {
	convey.Convey("AddTimeLock", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddTimeLock(c, 111)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddTimesRecord(t *testing.T) {
	convey.Convey("AddTimesRecord", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddTimesRecord(c, 111, "20200312")
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetAddTimesRecord(t *testing.T) {
	convey.Convey("GetAddTimesRecord", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GetAddTimesRecord(c, 111, "20200312")
			convCtx.Convey("Then err should be nil..", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldEqual, "True")
			})
		})
	})
}
