package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/ugc-season/service/api"
)

func TestStat(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("Stat", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			res, err := s.Stat(c, sid)
			fmt.Printf("%+v", res)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestStats(t *testing.T) {
	var (
		c   = context.TODO()
		sid = []int64{1}
	)
	convey.Convey("Stats", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, err := s.Stats(c, sid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSeason(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("Season", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, err := s.Season(c, sid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestView(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("View", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, err := s.View(c, sid)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpSeasonCache(t *testing.T) {
	var (
		c      = context.TODO()
		sid    = int64(1)
		action = "update"
	)
	convey.Convey("UpSeasonCache", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			err := s.UpSeasonCache(c, sid, action)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpperSeason(t *testing.T) {
	var (
		c   = context.TODO()
		req = &api.UpperListRequest{Mid: 100, PageNum: 1, PageSize: 10}
	)
	convey.Convey("UpperSeason", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			season, tcnt, tpage, err := s.UpperSeason(c, req)
			fmt.Printf("%+v", season)
			fmt.Println(tcnt)
			fmt.Println(tpage)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestViews(t *testing.T) {
	var (
		c   = context.TODO()
		sid = []int64{784}
	)
	convey.Convey("Views", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, err := s.Views(c, sid, 4)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
