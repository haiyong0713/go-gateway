package steins

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	gnk "gopkg.in/h2non/gock.v1"
)

func TestBnjSendWechat(t *testing.T) {
	convey.Convey("SendWechat", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			title = "互动视频新剧情树消息"
			msg   = "aid:57939708  查看链接:http://api.bilibili.com/x/stein/manager?aid=57939708"
			user  = "wuhao02"
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			defer gnk.OffAll()
			httpMock("POST", "").Reply(200).JSON(`{"RetCode":0}`)
			err := d.SendWechat(c, title, msg, user)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
