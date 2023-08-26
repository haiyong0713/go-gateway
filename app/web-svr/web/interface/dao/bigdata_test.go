package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestDaoRanking(t *testing.T) {
	convey.Convey("Ranking", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			rid      = int16(0)
			rankType = int(0)
			day      = int(0)
			arcType  = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.c.Host.Rank+_rankURL).Reply(200).JSON(`{"code":0,"num":8,"list":[{"aid":33986715,"score":704},{"aid":33913315,"score":571}]}`)
			res, err := d.Ranking(c, rid, rankType, day, arcType)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaorankURI(t *testing.T) {
	convey.Convey("rankURI", t, func(ctx convey.C) {
		var (
			rid      = int16(0)
			rankType = ""
			day      = int(0)
			arcType  = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := rankURI(rid, rankType, day, arcType)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestHotLabel(t *testing.T) {
	convey.Convey("HotLabel", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.hotLabelURL).Reply(200).JSON(`{"code":0,"list":[{"aid":111,"score":10},{"aid":2222,"score":20}]}`)
			aids, err := d.HotLabel(c)
			ctx.Convey("Then err should be nil.aids should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(aids, convey.ShouldNotBeNil)
			})
		})
	})
}
