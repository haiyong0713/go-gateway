package native

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeRawNativeModules(t *testing.T) {
	convey.Convey("RawNativeModules", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativeModules(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeRawSortModules(t *testing.T) {
	convey.Convey("RawSortModules", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			nat = int64(108)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawSortModules(c, nat, 1)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// RawNativeUkey
func TestRawNativeUkey(t *testing.T) {
	convey.Convey("RawNativeUkey", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			pid  = int64(4)
			ukey = "12345"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			red, err := d.RawNativeUkey(c, pid, ukey)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(red)
			})
		})
	})
}
