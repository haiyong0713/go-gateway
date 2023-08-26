package native

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeDynamics(t *testing.T) {
	convey.Convey("Dynamics", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{107}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.Dynamics(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeRawNativeDynamics(t *testing.T) {
	convey.Convey("RawNativeDynamics", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{108, 109}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativeDynamics(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
