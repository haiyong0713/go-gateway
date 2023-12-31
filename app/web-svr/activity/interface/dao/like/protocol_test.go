package like

import (
	"context"
	"testing"

	"fmt"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawActSubjectProtocol(t *testing.T) {
	convey.Convey("RawLikeMissionBuff", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10301)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RawActSubjectProtocol(c, sid)
			ctx.Convey("Then err should be nil.ID should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

func TestRawActSubjectProtocols(t *testing.T) {
	convey.Convey("RawLikeMissionBuff", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			sids = []int64{10554, 103553}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RawActSubjectProtocols(c, sids)
			ctx.Convey("Then err should be nil.ID should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}
