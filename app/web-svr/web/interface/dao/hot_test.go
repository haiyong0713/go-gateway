package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_HotAiRcmd(t *testing.T) {
	convey.Convey("HotAiRcmd", t, func(ctx convey.C) {
		var (
			mid    = int64(2)
			buvid  = ""
			pageNo = 0
			count  = 5
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, userFeature, code, err := d.HotAiRcmd(context.Background(), mid, buvid, pageNo, count)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.Printf("%+v", data)
				ctx.Printf("%+v", code)
				ctx.Printf("%+v", userFeature)
			})
		})
	})
}
