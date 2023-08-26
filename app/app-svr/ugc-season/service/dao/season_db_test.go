package dao

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/ugc-season/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestSeasonInfo(t *testing.T) {
	convey.Convey("TestSeasonInfo", t, func(ctx convey.C) {
		var (
			c         = context.Background()
			seasonID  = int64(1)
			seasonIDs = []int64{877, 785}
			season    *api.Season
			sections  []*api.Section
			episodes  map[int64][]*api.Episode
			stat      *api.Stat
			stats     map[int64]*api.Stat
			err       error
		)
		ctx.Convey("TestSeason", func(ctx convey.C) {
			season, err = d.SeasonInfo(c, seasonID)
			convey.Println(season)
			convey.Println(err)
		})
		ctx.Convey("TestSections", func(ctx convey.C) {
			sections, err = d.SectionsInfo(c, seasonID)
			convey.Println(sections)
			convey.Println(err)
		})
		ctx.Convey("TestEpisodes", func(ctx convey.C) {
			episodes, err = d.EpisodesInfo(c, seasonID)
			convey.Println(episodes)
			convey.Println(err)
		})
		ctx.Convey("TestStat", func(ctx convey.C) {
			stat, err = d.StatInfo(c, seasonID)
			convey.Println(stat)
			convey.Println(err)
		})
		ctx.Convey("TestStats", func(ctx convey.C) {
			stats, err = d.StatsInfo(c, seasonIDs)
			convey.Println(stats)
			convey.Println(err)
		})
	})
}

func TestUpperSeasonInfo(t *testing.T) {
	convey.Convey("UpperSeasonInfo", t, func(ctx convey.C) {
		var (
			c   = context.TODO()
			mid = int64(1)
		)
		ctx.Convey("UpperSeasonInfo", func(ctx convey.C) {
			sid, ptime, err := d.UpperSeasonInfo(c, mid)
			convey.Printf("%+v", sid)
			convey.Printf("%+v", ptime)
			convey.Printf("%+v", err)
		})
	})
}
