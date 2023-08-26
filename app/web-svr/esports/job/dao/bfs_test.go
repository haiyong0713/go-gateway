package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoBfsUpload(t *testing.T) {
	convey.Convey("BfsUpload", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			bs       = []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 2, 88}
			fileName = "test"
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			location, err := d.BfsUpload(c, bs, fileName)
			ctx.Convey("Then err should be nil.location should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(location, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoBfsPicture(t *testing.T) {
	convey.Convey("BfsPicture", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			url = "http://uat-i0.hdslb.com/bfs/esport/img/team/image/3364/RNG_Logo.png"
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			rs, err := d.ThirdGet(c, url)
			ctx.Convey("Then err should be nil.bs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(len(rs), convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}
