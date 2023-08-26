package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_TppShowInfo(t *testing.T) {
	convey.Convey("TppShowInfo", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			showID = int64(1275392)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			totalView, wantCnt, _ := d.TppShowInfo(c, showID)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.Printf("%d", totalView)
				ctx.Printf("%d", wantCnt)
			})
		})
	})
}

func TestDao_QQShowInfo(t *testing.T) {
	convey.Convey("QQShowInfo", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			totalView, err := d.QQShowInfo(c, "https://v.qq.com/x/cover/98frcd8mursfnqb/y0825q527dz.html")
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(totalView, convey.ShouldNotBeNil)
				ctx.Printf("%d", totalView)
			})
		})
	})
}
