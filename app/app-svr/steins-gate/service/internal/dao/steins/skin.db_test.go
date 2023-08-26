package steins

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawSkinList(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("nodeKey", t, func(ctx convey.C) {
		list, err := d.RawSkinList(c)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(list, convey.ShouldNotBeNil)
		})
	})
}
