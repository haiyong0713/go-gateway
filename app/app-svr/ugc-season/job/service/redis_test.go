package service

import (
	"context"
	"fmt"
	"testing"

	"go-common/library/time"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddUpperSeason(t *testing.T) {
	var (
		c     = context.TODO()
		sid   = int64(1)
		mid   = int64(1)
		ptime = time.Time(1560841403)
	)
	convey.Convey("AddUpperSeason", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			err := s.AddUpperSeason(c, sid, mid, ptime)
			fmt.Print(mid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDelUpperSeason(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
		mid = int64(1)
	)
	convey.Convey("DelUpperSeason", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			err := s.DelUpperSeason(c, sid, mid)
			fmt.Print(mid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
