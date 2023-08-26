package push

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestPushsignature(t *testing.T) {
	var (
		params map[string]string
		secret = "test"
	)
	convey.Convey("signature", t, func(ctx convey.C) {
		p1 := signature(params, secret)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}
