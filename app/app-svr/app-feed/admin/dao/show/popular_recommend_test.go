package show

import (
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestShowPopRecommendAdd(t *testing.T) {
	convey.Convey("PopRecommendAdd", t, func(ctx convey.C) {
		var (
			param = &show.PopRecommendAP{
				CardValue: "1",
				Label:     "test",
				Person:    "quguolin",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopRecommendAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowPopRecommendUpdate(t *testing.T) {
	convey.Convey("PopRecommendUpdate", t, func(ctx convey.C) {
		var (
			param = &show.PopRecommendUP{
				ID:    2,
				Label: "test111",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopRecommendUpdate(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowPopRecommendDelete(t *testing.T) {
	convey.Convey("PopRecommendDelete", t, func(ctx convey.C) {
		var (
			id = int64(200)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopRecommendDelete(id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowPopRFindByID(t *testing.T) {
	convey.Convey("PopRFindByID", t, func(ctx convey.C) {
		var (
			id = ""
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			_, err := d.PopRFindByID(id)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
