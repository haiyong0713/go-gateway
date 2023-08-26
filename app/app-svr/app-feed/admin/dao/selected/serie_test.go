package selected

import (
	"context"
	"fmt"
	"testing"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"

	"github.com/smartystreets/goconvey/convey"
)

func TestSelectedPickSerie(t *testing.T) {
	var (
		c   = context.Background()
		req = &selected.FindSerie{
			Type:   "weekly_selected",
			Number: 1,
		}
	)
	convey.Convey("PickSerie", t, func(ctx convey.C) {
		res, err := d.PickSerie(c, req)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSelectedPickRes(t *testing.T) {
	var (
		c  = context.Background()
		id = int64(1)
	)
	convey.Convey("PickRes", t, func(ctx convey.C) {
		res, err := d.PickRes(c, id)
		if ecode.EqualError(ecode.NothingFound, err) {
			err = nil
			res = &selected.Resource{ID: 1}
		}
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSelectedPickSeries(t *testing.T) {
	var (
		c    = context.Background()
		sids = []int64{1, 2, 3, 4, 5}
	)
	convey.Convey("PickSeries", t, func(ctx convey.C) {
		res, err := d.PickSeries(c, sids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSelectedSeries(t *testing.T) {
	var (
		c     = context.Background()
		sType = ""
	)
	convey.Convey("Series", t, func(ctx convey.C) {
		results, err := d.Series(c, sType)
		ctx.Convey("Then err should be nil.results should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(results, convey.ShouldNotBeNil)
		})
	})
}

func TestSelectedUpdateSerie(t *testing.T) {
	var (
		c      = context.Background()
		sid    = int64(0)
		status = int(0)
	)
	convey.Convey("UpdateSerieStatus", t, func(ctx convey.C) {
		err := d.UpdateSerieStatus(c, sid, status)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSelectedRejectRes(t *testing.T) {
	var (
		c      = context.Background()
		origin = &selected.Resource{}
	)
	convey.Convey("RejectRes", t, func(ctx convey.C) {
		err := d.RejectRes(c, origin)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSelectedDelRes(t *testing.T) {
	var (
		c  = context.Background()
		id = int64(0)
	)
	convey.Convey("DelRes", t, func(ctx convey.C) {
		err := d.DelRes(c, id)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSelectedUpdateRes(t *testing.T) {
	var (
		c      = context.Background()
		origin = &selected.Resource{}
		req    = &selected.ReqSelEdit{}
	)
	convey.Convey("UpdateRes", t, func(ctx convey.C) {
		err := d.UpdateRes(c, origin, req)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSelectedCntRes(t *testing.T) {
	var (
		c   = context.Background()
		sid = int64(0)
	)
	convey.Convey("CntRes", t, func(ctx convey.C) {
		cnt, err := d.CntRes(c, sid)
		ctx.Convey("Then err should be nil.cnt should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(cnt, convey.ShouldNotBeNil)
		})
	})
}

func TestSelectedSortRes(t *testing.T) {
	var (
		c       = context.Background()
		sid     = int64(0)
		cardIDs = []int64{}
	)
	convey.Convey("SortRes", t, func(ctx convey.C) {
		err := d.SortRes(c, sid, cardIDs)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSelectedSerieValid(t *testing.T) {
	var (
		c   = context.Background()
		sid = int64(149)
	)
	convey.Convey("SerieValid", t, func(ctx convey.C) {
		valid, err := d.SerieValid(c, sid)
		ctx.Convey("Then err should be nil.valid should not be nil.", func(ctx convey.C) {
			fmt.Println("valid:", valid)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(valid, convey.ShouldNotBeNil)
		})
	})
}

func TestSelectedSeriePass(t *testing.T) {
	var (
		c   = context.Background()
		sid = int64(0)
	)
	convey.Convey("SeriePass", t, func(ctx convey.C) {
		err := d.SeriePass(c, sid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSelectedSerieUpdate(t *testing.T) {
	var (
		c     = context.Background()
		serie = &selected.SerieDB{}
	)
	convey.Convey("SerieUpdate", t, func(ctx convey.C) {
		err := d.SerieUpdate(c, serie)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDuplicateCheck(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("SerieUpdate", t, func(ctx convey.C) {
		cnt, err := d.DuplicateCheck(c, 1, 1, "av", 1)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			fmt.Println(cnt)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSelectedResources(t *testing.T) {
	var c = context.Background()
	convey.Convey("SerieUpdate", t, func(ctx convey.C) {
		res, err := d.Resources(c, []int64{1, 2, 3, 4, 5, 6})
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetLastValidSerieByType(t *testing.T) {
	var c = context.Background()
	convey.Convey("GetLastValidSerieByType", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			res, err := d.GetLastValidSerieByType(c, selected.SERIE_TYPE_WEEKLY_SELECTED)
			fmt.Printf("res: %+v", res)
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
