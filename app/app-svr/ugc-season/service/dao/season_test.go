package dao

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/ugc-season/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestSeason(t *testing.T) {
	convey.Convey("Season", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			seasonID = int64(1)
			view     *api.View
			season   *api.Season
			err      error
		)
		ctx.Convey("TestSeason", func(ctx convey.C) {
			season, err = d.Season(c, seasonID)
			d.AddSeasonCache(c, season)
			ctx.Println(season)
			ctx.Println(err)
		})
		ctx.Convey("TestView", func(ctx convey.C) {
			view, err = d.View(c, seasonID)
			convey.Println(view)
			convey.Println(err)
		})
	})
}

func TestUpperSeason(t *testing.T) {
	var (
		c   = context.TODO()
		req = &api.UpperListRequest{Mid: 1, PageNum: 1, PageSize: 10}
	)
	convey.Convey("UpperSeason", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			season, tcnt, tpage, err := d.UpperSeason(c, req)
			fmt.Printf("%+v", season)
			fmt.Println(tcnt)
			fmt.Println(tpage)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
