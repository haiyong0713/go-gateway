package steins

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaotreeKey(t *testing.T) {
	var (
		aid = int64(3333)
	)
	convey.Convey("graphKey", t, func(ctx convey.C) {
		p1 := graphKey(aid)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestDaographCache(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10113629)
	)
	convey.Convey("graphCache", t, func(ctx convey.C) {
		a, err := d.graphCache(c, aid)
		fmt.Println(a)
		fmt.Println(err)
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(a, convey.ShouldNotBeNil)
		})
	})
}
