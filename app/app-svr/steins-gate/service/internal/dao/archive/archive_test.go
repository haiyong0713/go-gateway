package archive

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoArcView(t *testing.T) {
	var (
		ctx = context.Background()
		aid = int64(3333)
	)
	convey.Convey("ArcView", t, func(c convey.C) {
		res, err := d.ArcView(ctx, aid)
		c.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoArcs(t *testing.T) {
	var (
		ctx  = context.Background()
		aids = []int64{10098214, 10098215, 10098217}
	)
	convey.Convey("ArcView", t, func(c convey.C) {
		res, err := d.Arcs(ctx, aids)
		fmt.Println(res)
		c.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
