package message

import (
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/conf"

	"github.com/smartystreets/goconvey/convey"
)

func TestMessageNotify(t *testing.T) {
	convey.Convey("Notify", t, func(ctx convey.C) {
		var (
			mids = []int64{2089809}
			c    = &conf.MesConfig{
				MC:    "1_22_1",
				Title: "您的稿件被推荐啦",
				Msg:   "您的稿件【《稿件标题》】已被选入移动端首页推荐（2019-03-06 更新），期待您创作更加优秀的新作品~",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.Notify(mids, c)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestMessageNotifyTianma(t *testing.T) {
	convey.Convey("NotifyTianma", t, func(ctx convey.C) {
		var (
			mids  = []int64{27515432}
			typ   = "直播"
			param = "直播测试标题"
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.NotifyTianma(mids, typ, param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestMessageNotifyPopular(t *testing.T) {
	convey.Convey("NotifyPopular", t, func(ctx convey.C) {
		var (
			mids = []int64{27515432}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.NotifyPopular(mids)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
