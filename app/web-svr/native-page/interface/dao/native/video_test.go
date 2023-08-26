package native

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeNatVideos(t *testing.T) {
	convey.Convey("NatVideos", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NatVideos(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeRawNativeVideos(t *testing.T) {
	convey.Convey("RawNativeVideos", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativeVideos(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
