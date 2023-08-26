package card

import (
	"context"
	"go-gateway/app/app-svr/app-show/interface/model/selected"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCardoneSerieKey(t *testing.T) {
	var (
		number = int64(0)
		sType  = "weekly_selected"
	)
	convey.Convey("oneSerieKey", t, func(ctx convey.C) {
		p1 := oneSerieKey(number, sType)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestCardallSeriesKey(t *testing.T) {
	var (
		sType = "weekly_selected"
	)
	convey.Convey("allSeriesKey", t, func(ctx convey.C) {
		p1 := allSeriesKey(sType)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestCardSetAllSeries(t *testing.T) {
	var (
		c     = context.Background()
		sType = "weekly_selected"
		list  = []*selected.SerieFilter{}
	)
	convey.Convey("SetAllSeries", t, func(ctx convey.C) {
		list = append(list, &selected.SerieFilter{
			SerieCore: selected.SerieCore{
				Type:   "weekly_selected",
				Number: 1,
			},
		})
		err := d.SetAllSeries(c, sType, list)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCardAllSeriesCache(t *testing.T) {
	var (
		c     = context.Background()
		sType = "weekly_selected"
	)
	convey.Convey("AllSeriesCache", t, func(ctx convey.C) {
		res, err := d.AllSeriesCache(c, sType)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestCardAddSerieCache(t *testing.T) {
	var (
		c     = context.Background()
		serie = &selected.SerieFull{
			Config: &selected.SerieConfig{
				SerieCore: selected.SerieCore{
					Type:   "weekly_selected",
					Number: 1,
				},
			},
		}
	)
	convey.Convey("AddSerieCache", t, func(ctx convey.C) {
		err := d.AddSerieCache(c, serie)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCardPickSerieCache(t *testing.T) {
	var (
		c      = context.Background()
		sType  = "weekly_selected"
		number = int64(1)
	)
	convey.Convey("PickSerieCache", t, func(ctx convey.C) {
		serie, err := d.PickSerieCache(c, sType, number)
		ctx.Convey("Then err should be nil.serie should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(serie, convey.ShouldNotBeNil)
		})
	})
}

func TestBatchPickSerieCache(t *testing.T) {
	var (
		c       = context.Background()
		sType   = "weekly_selected"
		numbers = []int64{1, 2, 3}
	)
	convey.Convey("BatchPickSerieCache", t, func(ctx convey.C) {
		serie, err := d.BatchPickSerieCache(c, sType, numbers)
		ctx.Convey("Then err should be nil.serie should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(serie, convey.ShouldNotBeNil)
		})
	})
}
