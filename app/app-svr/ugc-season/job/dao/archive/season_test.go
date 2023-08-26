package archive

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSeason(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("Season", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, err := d.Season(c, sid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSections(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("Sections", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, err := d.Sections(c, sid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestEpisodes(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("Episodes", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, err := d.Episodes(c, sid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSeasonMaxPtime(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{2, 5, 7, 9, 10, 11}
	)
	convey.Convey("SeasonMaxPtime", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ptime, err := d.SeasonMaxPtime(c, aids)
			fmt.Print(ptime)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
