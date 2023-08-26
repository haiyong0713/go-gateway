package show

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-job/job/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestShowMerakNotify(t *testing.T) {
	var (
		ctx   = context.Background()
		merak = &show.Merak{
			Names:    []string{"zhaoshichen"},
			Template: "testTem",
			Title:    "testTitle",
		}
	)
	convey.Convey("MerakNotify", t, func(c convey.C) {
		httpMock("POST", d.conf.WechatAlert.Host).Reply(200).JSON("{}")
		err := d.MerakNotify(ctx, merak)
		c.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestShowMerakSign(t *testing.T) {
	var (
		params map[string]string
		secret = ""
	)
	convey.Convey("MerakSign", t, func(ctx convey.C) {
		p1, err := MerakSign(params, secret)
		ctx.Convey("Then err should be nil.p1 should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}
