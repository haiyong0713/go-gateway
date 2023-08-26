package dao

import (
	"context"
	"fmt"
	"testing"

	"go-common/library/time"

	"github.com/smartystreets/goconvey/convey"
)

func TestUpperSeasonCache(t *testing.T) {
	var (
		c     = context.TODO()
		mid   = int64(1)
		start = int64(1)
		end   = int64(10)
	)
	convey.Convey("UpperSeason", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			sids, total, err := d.UpperSeasonCache(c, mid, start, end)
			fmt.Printf("%+v", sids)
			fmt.Println(total)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddUpperSeasonCache(t *testing.T) {
	var (
		c     = context.TODO()
		mid   = int64(1)
		sid   = []int64{1}
		ptime = []time.Time{1560841403}
	)
	convey.Convey("AddUpperSeason", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			err := d.AddUpperSeasonCache(c, mid, sid, ptime)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSetUpperNoSeasonCache(t *testing.T) {
	var (
		c   = context.TODO()
		mid = int64(1)
	)
	convey.Convey("SetUpperNoSeasonCache", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			err := d.SetUpperNoSeasonCache(c, mid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
