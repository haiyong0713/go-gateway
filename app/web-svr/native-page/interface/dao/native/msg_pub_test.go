package native

import (
	"context"
	v1 "go-gateway/app/web-svr/native-page/interface/api"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeSendMsg(t *testing.T) {
	convey.Convey("SendMsg", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			page   = &v1.NativePage{ID: 1}
			onLine bool
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SendMsg(c, page, onLine)
			err = nil
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
