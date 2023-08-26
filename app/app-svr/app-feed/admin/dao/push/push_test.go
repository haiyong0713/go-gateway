package push

import (
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestPushNoticeUser(t *testing.T) {
	var (
		mids = []int64{}
		uuid = "uuid"
		body = &selected.Serie{
			PushTitle:    "title",
			PushSubtitle: "body",
		}
	)
	convey.Convey("NoticeUser", t, func(ctx convey.C) {
		httpMock("POST", d.pushURL).Reply(200).JSON(`{"code": 0,"data":1}`)
		err := d.NoticeUser(mids, uuid, body)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

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
