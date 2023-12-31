package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"go-gateway/app/web-svr/esports/interface/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoSearch(t *testing.T) {
	var (
		c     = context.Background()
		mid   = int64(111)
		p     = &model.ParamSearch{Keyword: "test", Pn: 1, Ps: 100, Sort: 0}
		buvid = ""
	)
	convey.Convey("Search", t, func(ctx convey.C) {
		rs, err := d.Search(c, mid, p, buvid)
		ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			println(rs)
		})
	})
}

func TestDaoSearchVideo(t *testing.T) {
	var (
		c = context.Background()
		p = &model.ParamVideo{Pn: 1, Ps: 30}
	)
	convey.Convey("SearchVideo", t, func(ctx convey.C) {
		rs, total, err := d.SearchVideo(c, p)
		ctx.Convey("Then err should be nil.rs,total should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(total, convey.ShouldNotBeNil)
			ctx.So(len(rs), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestDaoSearchContest(t *testing.T) {
	var (
		c = context.Background()
		p = &model.ParamContest{Pn: 1, Ps: 30}
	)
	convey.Convey("SearchContest", t, func(ctx convey.C) {
		rs, total, err := d.SearchContest(c, p)
		ctx.Convey("Then err should be nil.rs,total should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(total, convey.ShouldNotBeNil)
			ctx.So(len(rs), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestDaoFilterVideo(t *testing.T) {
	var (
		c = context.Background()
		p = &model.ParamFilter{}
	)
	convey.Convey("FilterVideo", t, func(ctx convey.C) {
		rs, err := d.FilterVideo(c, p)
		ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(rs, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoFilterMatch(t *testing.T) {
	var (
		c = context.Background()
		p = &model.ParamFilter{}
	)
	convey.Convey("FilterMatch", t, func(ctx convey.C) {
		rs, err := d.FilterMatch(c, p)
		ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(rs, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoFilterCale(t *testing.T) {
	var (
		c = context.Background()
		p = &model.ParamFilter{}
	)
	convey.Convey("FilterCale", t, func(ctx convey.C) {
		rs, err := d.FilterCale(c, p)
		ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(rs), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestDaoSearchFav(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(111)
		p   = &model.ParamFav{Pn: 1, Ps: 30}
	)
	convey.Convey("SearchFav", t, func(ctx convey.C) {
		rs, total, err := d.SearchFav(c, mid, p)
		ctx.Convey("Then err should be nil.rs,total should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(total, convey.ShouldNotBeNil)
			ctx.So(len(rs), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestDaoSeasonFav(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(111)
		p   = &model.ParamSeason{Pn: 1, Ps: 30}
	)
	convey.Convey("SeasonFav", t, func(ctx convey.C) {
		rs, total, err := d.SeasonFav(c, mid, p)
		ctx.Convey("Then err should be nil.rs,total should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(total, convey.ShouldNotBeNil)
			ctx.So(len(rs), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestDaoStimeFav(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(111)
		p   = &model.ParamSeason{Pn: 1, Ps: 30}
	)
	convey.Convey("StimeFav", t, func(ctx convey.C) {
		rs, total, err := d.StimeFav(c, mid, p)
		ctx.Convey("Then err should be nil.rs,total should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(total, convey.ShouldNotBeNil)
			ctx.So(len(rs), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestEsContestGuessRecent(t *testing.T) {
	var (
		c = context.Background()
		p = &model.ParamEsGuess{
			HomeID: 2,
			AwayID: 3,
			CID:    1,
		}
	)
	convey.Convey("EsContestGuessRecent", t, func(ctx convey.C) {
		contest, err := d.EsContestGuessRecent(c, p)
		ctx.Convey("Then err should be nil.rs,total should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		bs, _ := json.Marshal(contest)
		fmt.Println(string(bs))
	})
}

func TestEsHttpGet(t *testing.T) {
	var (
		c = context.Background()
	)
	params := url.Values{}
	params.Set("sql", "select id from esports_contests where id = 1")
	convey.Convey("EsContestGuessRecent", t, func(ctx convey.C) {
		contest, err := d.EsHTTPGet(c, params)
		ctx.Convey("Then err should be nil.rs,total should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		bs, _ := json.Marshal(contest)
		fmt.Println(string(bs))
	})
}
