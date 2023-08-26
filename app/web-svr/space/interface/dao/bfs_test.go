package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDecodeImageSize(t *testing.T) {
	convey.Convey("DecodeImageSize", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			x, y, name, err := d.DecodeImageSize(c, "http://i0.hdslb.com/bfs/archive/85bcb7691d595d0d85af80bc11804f746f025c07.jpg")
			ctx.Convey("Then err should be nil.blacklist should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(x, convey.ShouldNotBeNil)
				ctx.So(y, convey.ShouldNotBeNil)
				ctx.Printf("x:%d-y:%d name:%s", x, y, name)
			})
		})
	})
}
