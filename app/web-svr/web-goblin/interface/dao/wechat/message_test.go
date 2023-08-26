package wechat

import (
	"context"
	"encoding/json"
	"testing"

	"go-gateway/app/web-svr/web-goblin/interface/model/wechat"

	"github.com/smartystreets/goconvey/convey"
)

func TestSendMessage(t *testing.T) {
	convey.Convey("SendMessage", t, func(ctx convey.C) {
		var (
			c           = context.Background()
			accessToken = ""
		)
		sendArg := &wechat.SendMsg{
			Touser:  "",
			Msgtype: "link",
			Link: &wechat.LinkMsg{
				Title:       "点击下载bilibili",
				Description: "哔哩哔哩干杯",
				URL:         "https://www.bilibili.com",
				ThumbURL:    "",
			},
		}
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			arg, _ := json.Marshal(sendArg)
			err := d.SendMessage(c, accessToken, arg)
			ctx.Convey("Then err should be nil.qrcode should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}
