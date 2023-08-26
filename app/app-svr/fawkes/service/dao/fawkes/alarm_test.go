package fawkes

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/model"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
)

// 测试webhook推送
func TestWebhookPush(t *testing.T) {
	convey.Convey("WebHook", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			params := &cimdl.HookParam{
				AppKey:      "bstar_pogo_a64",
				AppName:     "POGO国际版-Android64",
				BuildID:     382921,
				GitlabJobID: 8963085,
				CTime:       time.Now().Unix(),
				PackURL:     "/mnt/build-archive/archive/fawkes/pack/bstar_pogo_a64/8963085/pogo-debug.apk",
			}
			hook := &model.Hook{
				URI:    "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=71451bd8-3ecc-45ae-881b-f9fc48864c3a",
				Method: "GET",
			}
			err := d.Hook(context.Background(), params, hook)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
