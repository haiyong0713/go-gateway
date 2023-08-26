package dao

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_Roles(t *testing.T) {
	convey.Convey("Roles", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			rs := d.Roles(_lolGame)
			convCtx.Convey("Then err should be nil.rs should not be nil.", func(convCtx convey.C) {
				convCtx.So(len(rs), convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}
