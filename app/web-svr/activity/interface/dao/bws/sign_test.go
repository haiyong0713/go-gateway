package bws

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawBwsSign(t *testing.T) {
	convey.Convey("RawBwsSign", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = []int64{1, 2, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawBwsSign(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestRawSigns(t *testing.T) {
	convey.Convey("RawSigns", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			pid = int64(7)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawSigns(c, pid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestDescSignPoint(t *testing.T) {
	convey.Convey("DescSignPoint", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			pid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.IncrSignPoint(c, pid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
