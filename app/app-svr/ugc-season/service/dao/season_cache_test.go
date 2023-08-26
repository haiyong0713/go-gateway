package dao

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/ugc-season/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestSeasonCache(t *testing.T) {
	convey.Convey("TestSeasonCache", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			seasonID = int64(1)
			season   *api.Season
			stat     *api.Stat
			err      error
		)
		ctx.Convey("TestAddSeasonCache", func(ctx convey.C) {
			season, _ = d.SeasonInfo(c, seasonID)
			convey.Println(season)
			if season != nil {
				err = d.AddSeasonCache(c, season)
				convey.So(err, convey.ShouldBeNil)
			}
		})
		ctx.Convey("TestSeasonCache", func(ctx convey.C) {
			season, err = d.SeasonRdsCache(c, seasonID)
			convey.Println(season)
			convey.Println(err)
		})
		ctx.Convey("TestAddStatCache", func(ctx convey.C) {
			stat, err = d.StatInfo(c, seasonID)
			convey.Println(stat)
			if stat != nil {
				err = d.AddStCache(c, stat)
			}
		})
		ctx.Convey("TestStatCache", func(ctx convey.C) {
			stat, err = d.StCache(c, seasonID)
			convey.Println(stat)
			convey.Println(err)
		})
	})
}
