package show

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	xsql "go-common/library/database/sql"

	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
)

func TestShowSerieRecovery(t *testing.T) {
	var (
		c   = context.Background()
		sid = int64(0)
	)
	convey.Convey("SerieRecovery", t, func(ctx convey.C) {
		monkey.PatchInstanceMethod(reflect.TypeOf(d.db), "Exec", func(_ *xsql.DB, _ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
			return nil, nil
		})
		defer monkey.UnpatchAll()
		err := d.SerieRecovery(c, sid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestShowAICount(t *testing.T) {
	var (
		c   = context.Background()
		sid = int64(1)
	)
	convey.Convey("AICount", t, func(ctx convey.C) {
		count, err := d.AICount(c, sid)
		ctx.Convey("Then err should be nil.count should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(count, convey.ShouldNotBeNil)
		})
	})
}

func TestShowPickSerie(t *testing.T) {
	var (
		c     = context.Background()
		sType = "weekly_selected"
	)
	convey.Convey("PickSerie", t, func(ctx convey.C) {
		res, err := d.PickSerie(c, sType)
		if err == sql.ErrNoRows {
			return
		}
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestShowMaxPosition(t *testing.T) {
	var (
		c   = context.Background()
		sid = int64(11)
	)
	convey.Convey("MaxPosition", t, func(ctx convey.C) {
		max, err := d.MaxPosition(c, sid)
		ctx.Convey("Then err should be nil.max should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(max, convey.ShouldNotBeNil)
		})
	})
}

func TestShowMaxNumber(t *testing.T) {
	var (
		c     = context.Background()
		sType = "weekly_selected"
	)
	convey.Convey("MaxNumber", t, func(ctx convey.C) {
		number, err := d.MaxNumber(c, sType)
		ctx.Convey("Then err should be nil.number should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(number, convey.ShouldNotBeNil)
		})
	})
}

func TestShowRefreshSeries(t *testing.T) {
	var (
		c     = context.Background()
		sType = "weekly_selected"
	)
	convey.Convey("RefreshSeries", t, func(ctx convey.C) {
		err := d.RefreshSeries(c, sType)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestShowRefreshSerieMC(t *testing.T) {
	var (
		c      = context.Background()
		sType  = "weekly_selected"
		number = int64(1)
	)
	convey.Convey("RefreshSingleSerie", t, func(ctx convey.C) {
		err := d.RefreshSingleSerie(c, sType, number)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
