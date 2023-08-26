package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestTranSeason(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("tranSeason", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, _, _, _, _, mid, err := s.tranSeason(c, sid)
			fmt.Print(mid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
